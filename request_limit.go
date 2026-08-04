package gofofa

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
)

const advancedQueryInterval = time.Second

type queryRateLimiter struct {
	mu         sync.Mutex
	interval   time.Duration
	next       time.Time
	sharedPath string
}

func newQueryRateLimiter(interval time.Duration) *queryRateLimiter {
	return &queryRateLimiter{interval: interval}
}

func queryInterval(account AccountInfo) time.Duration {
	if account.IsVIP && account.VIPLevel == VipLevelAdvanced {
		return advancedQueryInterval
	}
	return 0
}

func (l *queryRateLimiter) wait(ctx context.Context) (time.Duration, error) {
	if l == nil || l.interval <= 0 {
		return 0, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if l.sharedPath != "" {
		return l.waitShared(ctx)
	}

	started := time.Now()

	for {
		l.mu.Lock()
		waitFor := time.Until(l.next)
		if waitFor <= 0 {
			l.next = time.Now().Add(l.interval)
			l.mu.Unlock()
			return 0, nil
		}
		l.mu.Unlock()

		timer := time.NewTimer(waitFor)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return time.Since(started), ctx.Err()
		}
	}
}

func (l *queryRateLimiter) waitShared(ctx context.Context) (time.Duration, error) {
	started := time.Now()
	waitFor, err := l.reserveSharedSlot(ctx)
	if err != nil {
		return time.Since(started), err
	}
	if waitFor <= 0 {
		return 0, nil
	}

	timer := time.NewTimer(waitFor)
	select {
	case <-timer.C:
		return time.Since(started), nil
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return time.Since(started), ctx.Err()
	}
}

func (l *queryRateLimiter) withQuerySlot(ctx context.Context, fn func() error) (time.Duration, error) {
	if l == nil || l.interval <= 0 {
		return 0, fn()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if l.sharedPath != "" {
		return l.withSharedQuerySlot(ctx, fn)
	}

	started := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := waitForRetry(ctx, time.Until(l.next)); err != nil {
		return time.Since(started), err
	}
	l.next = time.Now().Add(l.interval)
	return time.Since(started), fn()
}

func (l *queryRateLimiter) withSharedQuerySlot(ctx context.Context, fn func() error) (waited time.Duration, err error) {
	started := time.Now()
	lock := flock.New(l.sharedPath + ".lock")
	locked, err := lock.TryLockContext(ctx, 5*time.Millisecond)
	if err != nil {
		return 0, err
	}
	if !locked {
		return 0, fmt.Errorf("acquire shared rate limit lock: lock was not acquired")
	}
	defer func() {
		if unlockErr := lock.Unlock(); err == nil && unlockErr != nil {
			err = unlockErr
		}
	}()

	waitFor, err := l.reserveSharedSlotWithLock()
	if err != nil {
		return time.Since(started), err
	}
	if err = waitForRetry(ctx, waitFor); err != nil {
		return time.Since(started), err
	}
	return time.Since(started), fn()
}

func (l *queryRateLimiter) reserveSharedSlot(ctx context.Context) (waitFor time.Duration, err error) {
	lock := flock.New(l.sharedPath + ".lock")
	locked, err := lock.TryLockContext(ctx, 5*time.Millisecond)
	if err != nil {
		return 0, err
	}
	if !locked {
		return 0, fmt.Errorf("acquire shared rate limit lock: lock was not acquired")
	}
	defer func() {
		if unlockErr := lock.Unlock(); err == nil && unlockErr != nil {
			err = unlockErr
		}
	}()
	return l.reserveSharedSlotWithLock()
}

func (l *queryRateLimiter) reserveSharedSlotWithLock() (time.Duration, error) {
	now := time.Now()
	next := now
	data, err := os.ReadFile(l.sharedPath)
	if err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	if err == nil && len(strings.TrimSpace(string(data))) == 0 {
		return 0, fmt.Errorf("invalid shared rate limit state: empty state")
	}
	if len(data) > 0 {
		unixNano, parseErr := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("invalid shared rate limit state: %w", parseErr)
		}
		next = time.Unix(0, unixNano)
	}
	if next.Before(now) {
		next = now
	}
	if err := writeSharedRateLimitState(l.sharedPath, next.Add(l.interval)); err != nil {
		return 0, err
	}
	return time.Until(next), nil
}

func writeSharedRateLimitState(path string, next time.Time) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(path), ".gofofa-rate-limit-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		if removeErr := os.Remove(tempPath); removeErr != nil && !os.IsNotExist(removeErr) {
			if err == nil {
				err = removeErr
			}
		}
	}()

	if _, writeErr := temp.WriteString(strconv.FormatInt(next.UnixNano(), 10)); writeErr != nil {
		if closeErr := temp.Close(); closeErr != nil {
			return fmt.Errorf("write shared rate limit state: %w; close temporary file: %v", writeErr, closeErr)
		}
		return writeErr
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func isQueryAPI(apiURI string) bool {
	return strings.HasPrefix(apiURI, "search/")
}

func isRateLimitResponse(statusCode int, body []byte) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}

	var response struct {
		Error  bool   `json:"error"`
		Errmsg string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return false
	}
	return response.Error && strings.Contains(response.Errmsg, "45012")
}

func sharedRateLimitPath(server, key string) string {
	scope := sha256.Sum256([]byte(server + "\x00" + key))
	return filepath.Join(os.TempDir(), "gofofa-rate-limit-"+hex.EncodeToString(scope[:])+".state")
}

func retryAfterDuration(header string, fallback time.Duration) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(header)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(header); err == nil {
		if delay := time.Until(retryAt); delay > 0 {
			return delay
		}
		return 0
	}
	return fallback
}

func retryDelay(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		return 0
	}
	delay := base
	for i := 0; i < attempt; i++ {
		if delay >= 30*time.Second {
			return 30 * time.Second
		}
		delay *= 2
	}
	if delay > 30*time.Second {
		return 30 * time.Second
	}
	return delay
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
