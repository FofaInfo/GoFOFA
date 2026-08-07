package gofofa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/flock"
	"github.com/sirupsen/logrus"
)

func newRateLimitTestClient(server *httptest.Server) *Client {
	return &Client{
		Server:     server.URL,
		APIVersion: "v1",
		Account: AccountInfo{
			IsVIP:    true,
			VIPLevel: VipLevelAdvanced,
		},
		httpClient:          server.Client(),
		logger:              logrus.New(),
		ctx:                 context.Background(),
		queryLimiter:        newQueryRateLimiter(0),
		rateLimitRetries:    3,
		rateLimitRetryDelay: 0,
	}
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		panic(err)
	}
}

func searchRows(size int, value string) [][]string {
	rows := make([][]string, size)
	for i := range rows {
		rows[i] = []string{fmt.Sprintf("%s-%d", value, i)}
	}
	return rows
}

func TestHostSearchRateLimiterSpacesPages(t *testing.T) {
	var (
		mu        sync.Mutex
		requestAt []time.Time
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/all" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		mu.Lock()
		requestAt = append(requestAt, time.Now())
		mu.Unlock()

		page := r.URL.Query().Get("page")
		rows := searchRows(1, "last")
		if page == "1" {
			rows = searchRows(1000, "first")
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"error": false, "results": rows})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	client.queryLimiter = newQueryRateLimiter(20 * time.Millisecond)
	results, err := client.HostSearch("port=80", -1, []string{"ip"})
	if err != nil {
		t.Fatalf("HostSearch returned error: %v", err)
	}
	if len(results) != 1001 {
		t.Fatalf("result count = %d, want 1001", len(results))
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requestAt) != 2 {
		t.Fatalf("search request count = %d, want 2", len(requestAt))
	}
	if elapsed := requestAt[1].Sub(requestAt[0]); elapsed < 15*time.Millisecond {
		t.Fatalf("page interval = %s, want at least 15ms", elapsed)
	}
}

func TestHostSearchRateLimiterWaitsAfterSlowRequest(t *testing.T) {
	var (
		mu        sync.Mutex
		requestAt []time.Time
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestAt = append(requestAt, time.Now())
		requestNumber := len(requestAt)
		mu.Unlock()
		if requestNumber == 1 {
			time.Sleep(40 * time.Millisecond)
		}
		rows := [][]string{{"198.51.100.1"}}
		if requestNumber == 1 {
			rows = searchRows(1000, "first")
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"results": rows,
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	client.queryLimiter = newQueryRateLimiter(20 * time.Millisecond)
	if _, err := client.HostSearch("port=80", -1, []string{"ip"}); err != nil {
		t.Fatalf("HostSearch returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requestAt) != 2 {
		t.Fatalf("request count = %d, want two requests", len(requestAt))
	}
	if elapsed := requestAt[1].Sub(requestAt[0]); elapsed < 35*time.Millisecond {
		t.Fatalf("request interval = %s, want at least 35ms", elapsed)
	}
}

func TestHostSearchRetriesRateLimitResponses(t *testing.T) {
	for _, test := range []struct {
		name       string
		statusCode int
	}{
		{name: "business error", statusCode: http.StatusOK},
		{name: "http 429", statusCode: http.StatusTooManyRequests},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/search/all" {
					calls++
					if calls == 1 {
						writeJSON(w, test.statusCode, map[string]interface{}{
							"error":  true,
							"errmsg": "[45012] 请求速度过快",
						})
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"error": false, "results": [][]string{{"198.51.100.1"}},
					})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"error": false})
			}))
			defer server.Close()

			client := newRateLimitTestClient(server)
			client.rateLimitRetries = 1
			results, err := client.HostSearch("port=80", 1, []string{"ip"})
			if err != nil {
				t.Fatalf("HostSearch returned error: %v", err)
			}
			if len(results) != 1 || calls != 2 {
				t.Fatalf("results=%v calls=%d, want one result and two calls", results, calls)
			}
		})
	}
}

func TestHostSearchUsesErrorFlagAsAuthoritative(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/all" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		calls++
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"errmsg":  "[45012] informational text",
			"results": [][]string{{"198.51.100.1"}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	results, err := client.HostSearch("port=80", 1, []string{"ip"})
	if err != nil {
		t.Fatalf("HostSearch returned error: %v", err)
	}
	if calls != 1 || len(results) != 1 || results[0][0] != "198.51.100.1" {
		t.Fatalf("results=%v calls=%d, want one successful result and one call", results, calls)
	}
}

func TestHostSearchReturnsRateLimitAfterRetryLimit(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/search/all" {
			calls++
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"error": true, "errmsg": "[45012] 请求速度过快",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"error": false})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	client.rateLimitRetries = 2
	_, err := client.HostSearch("port=80", 1, []string{"ip"})
	if err == nil || !strings.Contains(err.Error(), "45012") || !strings.Contains(err.Error(), "after 2 retries") {
		t.Fatalf("error = %v, want final 45012 error with retry count", err)
	}
	if calls != 3 {
		t.Fatalf("search request count = %d, want 3", calls)
	}
}

func TestHostSearchReturnsErrorFlagWithoutMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   true,
			"results": [][]string{{"should-not-be-returned"}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	results, err := client.HostSearch("port=80", 1, []string{"ip"})
	if err == nil || !strings.Contains(err.Error(), "fofa search failed") {
		t.Fatalf("error = %v, want error flag failure", err)
	}
	if results != nil {
		t.Fatalf("results = %v, want nil on error response", results)
	}
}

func TestDumpSearchReturnsErrorFlagWithoutMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   true,
			"results": [][]string{{"should-not-be-delivered"}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	callbackCalls := 0
	err := client.DumpSearch("port=80", -1, 1, []string{"ip"}, func([][]string, int) error {
		callbackCalls++
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "fofa search failed") {
		t.Fatalf("error = %v, want error flag failure", err)
	}
	if callbackCalls != 0 {
		t.Fatalf("callback calls = %d, want no callback on error response", callbackCalls)
	}
}

func TestDumpSearchStopsWhenCursorDoesNotAdvance(t *testing.T) {
	calls := 0
	callbackCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/next" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		calls++
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error": false,
			"size":  2,
			"next":  "same-cursor",
			"results": [][]string{{
				"198.51.100.1",
			}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	err := client.DumpSearch("port=80", -1, 1, []string{"ip"}, func([][]string, int) error {
		callbackCalls++
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "cursor did not advance") {
		t.Fatalf("error = %v, want cursor progress error", err)
	}
	if calls != 2 {
		t.Fatalf("search/next request count = %d, want 2", calls)
	}
	if callbackCalls != 1 {
		t.Fatalf("callback count = %d, want only the non-stalled page", callbackCalls)
	}
}

func TestDumpSearchRejectsStalledPageThatWouldReachLimit(t *testing.T) {
	calls := 0
	var delivered [][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/next" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		calls++
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"size":    2,
			"next":    "same-cursor",
			"results": [][]string{{"198.51.100.1"}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	err := client.DumpSearch("port=80", 2, 1, []string{"ip"}, func(results [][]string, _ int) error {
		delivered = append(delivered, results...)
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "cursor did not advance") {
		t.Fatalf("error = %v, want cursor progress error", err)
	}
	if calls != 2 || len(delivered) != 1 {
		t.Fatalf("requests = %d, delivered = %v; want only the non-stalled page", calls, delivered)
	}
}

func TestDumpSearchDeliversAdvancingPageThatReachesLimit(t *testing.T) {
	calls := 0
	var delivered [][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"size":    2,
			"next":    fmt.Sprintf("cursor-%d", calls),
			"results": [][]string{{fmt.Sprintf("198.51.100.%d", calls)}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	err := client.DumpSearch("port=80", 2, 1, []string{"ip"}, func(results [][]string, _ int) error {
		delivered = append(delivered, results...)
		return nil
	})
	if err != nil {
		t.Fatalf("DumpSearch returned error with advancing cursor: %v", err)
	}
	if calls != 2 || len(delivered) != 2 {
		t.Fatalf("requests = %d, delivered = %v; want two requests and two results", calls, delivered)
	}
}

func TestDumpSearchLimitsFinalBatch(t *testing.T) {
	calls := 0
	var requestedSize string
	var delivered [][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		requestedSize = r.URL.Query().Get("size")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error": false,
			"size":  3,
			"results": [][]string{
				{"198.51.100.1"},
				{"198.51.100.2"},
				{"198.51.100.3"},
			},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	err := client.DumpSearch("port=80", 2, 3, []string{"ip"}, func(results [][]string, _ int) error {
		delivered = append(delivered, results...)
		return nil
	})
	if err != nil {
		t.Fatalf("DumpSearch returned error: %v", err)
	}
	if calls != 1 || requestedSize != "2" || len(delivered) != 2 {
		t.Fatalf("calls=%d requested_size=%q delivered=%v; want one request for two results", calls, requestedSize, delivered)
	}
}

func TestDumpSearchStopsWhenShortPageCursorDoesNotAdvance(t *testing.T) {
	calls := 0
	callbackCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/next" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		calls++
		results := [][]string{{"198.51.100.1"}, {"198.51.100.2"}}
		if calls == 2 {
			results = results[:1]
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"size":    3,
			"next":    "same-cursor",
			"results": results,
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	err := client.DumpSearch("port=80", -1, 2, []string{"ip"}, func([][]string, int) error {
		callbackCalls++
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "cursor did not advance") {
		t.Fatalf("error = %v, want cursor progress error", err)
	}
	if calls != 2 || callbackCalls != 1 {
		t.Fatalf("calls=%d callbacks=%d, want one delivered page", calls, callbackCalls)
	}
}

func TestConcurrentHostSearchesShareRateLimiter(t *testing.T) {
	var (
		mu        sync.Mutex
		requestAt []time.Time
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/all" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		mu.Lock()
		requestAt = append(requestAt, time.Now())
		mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"results": [][]string{{"198.51.100.1"}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	client.queryLimiter = newQueryRateLimiter(20 * time.Millisecond)

	var wg sync.WaitGroup
	errs := make(chan error, 3)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.HostSearch("port=80", 1, []string{"ip"})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("HostSearch returned error: %v", err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requestAt) != 3 {
		t.Fatalf("search request count = %d, want 3", len(requestAt))
	}
	sort.Slice(requestAt, func(i, j int) bool { return requestAt[i].Before(requestAt[j]) })
	for i := 1; i < len(requestAt); i++ {
		if elapsed := requestAt[i].Sub(requestAt[i-1]); elapsed < 15*time.Millisecond {
			t.Fatalf("request interval = %s, want at least 15ms", elapsed)
		}
	}
}

func TestSharedQueryRateLimiterReservesAcrossInstances(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	first := newQueryRateLimiter(20 * time.Millisecond)
	first.sharedPath = path
	second := newQueryRateLimiter(20 * time.Millisecond)
	second.sharedPath = path

	var requestAt []time.Time
	if _, err := first.withQuerySlot(context.Background(), func() error {
		requestAt = append(requestAt, time.Now())
		return nil
	}); err != nil {
		t.Fatalf("first request: %v", err)
	}
	if _, err := second.withQuerySlot(context.Background(), func() error {
		requestAt = append(requestAt, time.Now())
		return nil
	}); err != nil {
		t.Fatalf("second request: %v", err)
	}
	if elapsed := requestAt[1].Sub(requestAt[0]); elapsed < 15*time.Millisecond {
		t.Fatalf("shared request interval = %s, want at least 15ms", elapsed)
	}
}

func TestSharedQueryRateLimiterRecoversInvalidState(t *testing.T) {
	for name, contents := range map[string][]byte{
		"empty":   nil,
		"corrupt": []byte("corrupt"),
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "query-rate-limit.state")
			if err := os.WriteFile(path, contents, 0o600); err != nil {
				t.Fatalf("write invalid state: %v", err)
			}

			limiter := newQueryRateLimiter(20 * time.Millisecond)
			limiter.sharedPath = path
			called := false
			_, err := limiter.withQuerySlot(context.Background(), func() error {
				called = true
				return nil
			})
			if err != nil || !called {
				t.Fatalf("recovered request error = %v, called = %v", err, called)
			}

			rewritten, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read rewritten state: %v", err)
			}
			if _, err := strconv.ParseInt(strings.TrimSpace(string(rewritten)), 10, 64); err != nil {
				t.Fatalf("rewritten state %q is not a timestamp: %v", rewritten, err)
			}
		})
	}
}

// ========== queryInterval ==========

func TestQueryInterval_AdvancedVIP(t *testing.T) {
	account := AccountInfo{IsVIP: true, VIPLevel: VipLevelAdvanced}
	if d := queryInterval(account); d != advancedQueryInterval {
		t.Fatalf("queryInterval = %s, want %s", d, advancedQueryInterval)
	}
}

func TestQueryInterval_NonVIP(t *testing.T) {
	account := AccountInfo{IsVIP: false, VIPLevel: VipLevelAdvanced}
	if d := queryInterval(account); d != 0 {
		t.Fatalf("queryInterval = %s, want 0", d)
	}
}

func TestQueryInterval_NormalVIP(t *testing.T) {
	account := AccountInfo{IsVIP: true, VIPLevel: VipLevelNormal}
	if d := queryInterval(account); d != 0 {
		t.Fatalf("queryInterval = %s, want 0", d)
	}
}

// ========== withQuerySlot ==========

func TestWithQuerySlot_NilLimiter(t *testing.T) {
	var limiter *queryRateLimiter
	called := false
	waited, err := limiter.withQuerySlot(context.Background(), func() error {
		called = true
		return nil
	})
	if err != nil || !called || waited != 0 {
		t.Fatalf("nil limiter withQuerySlot = (%s, %v, called=%v), want (0, nil, true)", waited, err, called)
	}
}

func TestWithQuerySlot_ZeroInterval(t *testing.T) {
	limiter := newQueryRateLimiter(0)
	called := false
	waited, err := limiter.withQuerySlot(context.Background(), func() error {
		called = true
		return fmt.Errorf("fn error")
	})
	if err == nil || !called || waited != 0 {
		t.Fatalf("zero-interval withQuerySlot = (%s, %v, called=%v)", waited, err, called)
	}
}

func TestWithQuerySlot_NilContext(t *testing.T) {
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	called := false
	_, err := limiter.withQuerySlot(nil, func() error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("nil ctx withQuerySlot err = %v, called = %v", err, called)
	}
}

func TestWithQuerySlot_FnError(t *testing.T) {
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	_, err := limiter.withQuerySlot(context.Background(), func() error {
		return fmt.Errorf("fn error")
	})
	if err == nil || err.Error() != "fn error" {
		t.Fatalf("withQuerySlot err = %v, want 'fn error'", err)
	}
}

func TestWithQuerySlot_SharedPathRedirect(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	called := false
	_, err := limiter.withQuerySlot(context.Background(), func() error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("shared-path withQuerySlot err = %v, called = %v", err, called)
	}
}

func TestWithQuerySlot_CtxCancelled(t *testing.T) {
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.next = time.Now().Add(500 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := limiter.withQuerySlot(ctx, func() error { return nil })
	if err != context.Canceled {
		t.Fatalf("cancelled withQuerySlot err = %v, want context.Canceled", err)
	}
}

// ========== withSharedQuerySlot ==========

func TestWithSharedQuerySlot_HappyPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	called := false
	waited, err := limiter.withSharedQuerySlot(context.Background(), func() error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("withSharedQuerySlot err = %v, called = %v, waited = %s", err, called, waited)
	}
}

func TestWithSharedQuerySlot_FnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	_, err := limiter.withSharedQuerySlot(context.Background(), func() error {
		return fmt.Errorf("fn error")
	})
	if err == nil || err.Error() != "fn error" {
		t.Fatalf("withSharedQuerySlot err = %v, want 'fn error'", err)
	}
}

func TestWithSharedQuerySlotReleasesLockBeforeRequest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path

	_, err := limiter.withSharedQuerySlot(context.Background(), func() error {
		probe := flock.New(path + ".lock")
		locked, lockErr := probe.TryLock()
		if lockErr != nil {
			return lockErr
		}
		if !locked {
			return errors.New("shared lock remained held during request")
		}
		return probe.Unlock()
	})
	if err != nil {
		t.Fatalf("request callback could not acquire shared lock: %v", err)
	}
}

func TestWithSharedQuerySlotReturnsLockContention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	lockPath := path + ".lock"
	lock := flock.New(lockPath)
	locked, err := lock.TryLock()
	if err != nil {
		t.Fatalf("create lock: %v", err)
	}
	defer lock.Unlock()
	if !locked {
		t.Fatal("failed to acquire lock")
	}

	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = limiter.withSharedQuerySlot(ctx, func() error { return nil })
	if err == nil {
		t.Fatal("withSharedQuerySlot should fail when lock is held by another process")
	}
}

// ========== writeSharedRateLimitState error paths ==========

func TestWriteSharedRateLimitState_BadDir(t *testing.T) {
	err := writeSharedRateLimitState("/nonexistent/path/state.file", time.Now())
	if err == nil {
		t.Fatal("writeSharedRateLimitState should fail for nonexistent directory")
	}
}

// ========== isRateLimitResponse ==========

func TestIsRateLimitResponse_InvalidJSON(t *testing.T) {
	if isRateLimitResponse(http.StatusOK, []byte("not json")) {
		t.Fatal("isRateLimitResponse should return false for invalid JSON")
	}
}

func TestIsRateLimitResponse_ErrorFalse(t *testing.T) {
	if isRateLimitResponse(http.StatusOK, []byte(`{"error":false,"errmsg":"45012"}`)) {
		t.Fatal("isRateLimitResponse should return false when error=false")
	}
}

func TestIsRateLimitResponse_No45012(t *testing.T) {
	if isRateLimitResponse(http.StatusOK, []byte(`{"error":true,"errmsg":"other error"}`)) {
		t.Fatal("isRateLimitResponse should return false without 45012")
	}
}

func TestIsRateLimitResponseRequiresExactErrorCode(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "real FOFA rate limit response",
			body: `{"error":true,"errmsg":"[45012] 请求速度过快"}`,
			want: true,
		},
		{
			name: "different code containing the digits",
			body: `{"error":true,"errmsg":"[145012] another error"}`,
		},
		{
			name: "digits only in message",
			body: `{"error":true,"errmsg":"request 45012 failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRateLimitResponse(http.StatusOK, []byte(tt.body)); got != tt.want {
				t.Fatalf("isRateLimitResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRateLimitResponse_EmptyBody(t *testing.T) {
	if isRateLimitResponse(http.StatusOK, []byte{}) {
		t.Fatal("isRateLimitResponse should return false for empty body")
	}
}

// ========== sharedRateLimitPath ==========

func TestSharedRateLimitPath(t *testing.T) {
	path := sharedRateLimitPath("https://fofa.info", "secret-key")
	if !strings.Contains(path, "gofofa-rate-limit-") || !strings.HasSuffix(path, ".state") {
		t.Fatalf("sharedRateLimitPath = %s, want gofofa-rate-limit-*.state", path)
	}
	// same inputs should produce same path
	if p2 := sharedRateLimitPath("https://fofa.info", "secret-key"); p2 != path {
		t.Fatalf("sharedRateLimitPath not deterministic: %s vs %s", path, p2)
	}
	// different inputs should produce different path
	if p3 := sharedRateLimitPath("https://fofa.info", "other-key"); p3 == path {
		t.Fatalf("sharedRateLimitPath should differ for different keys")
	}
}

// ========== retryAfterDuration ==========

func TestRetryAfterDuration_Seconds(t *testing.T) {
	d := retryAfterDuration("5", time.Second)
	if d != 5*time.Second {
		t.Fatalf("retryAfterDuration = %s, want 5s", d)
	}
}

func TestRetryAfterDuration_CapsLargeSeconds(t *testing.T) {
	for _, header := range []string{"31", "9223372036854775807"} {
		if d := retryAfterDuration(header, time.Second); d != maxRateLimitRetryDelay {
			t.Fatalf("retryAfterDuration(%q) = %s, want %s", header, d, maxRateLimitRetryDelay)
		}
	}
}

func TestRetryAfterDuration_ZeroSeconds(t *testing.T) {
	d := retryAfterDuration("0", time.Second)
	if d != 0 {
		t.Fatalf("retryAfterDuration = %s, want 0", d)
	}
}

func TestRetryAfterDuration_HTTPDate(t *testing.T) {
	retryAt := time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat)
	d := retryAfterDuration(retryAt, time.Second)
	if d < 29*time.Second || d > 31*time.Second {
		t.Fatalf("retryAfterDuration = %s, want ~30s", d)
	}
}

func TestRetryAfterDuration_CapsFutureHTTPDate(t *testing.T) {
	retryAt := time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)
	if d := retryAfterDuration(retryAt, time.Second); d != maxRateLimitRetryDelay {
		t.Fatalf("retryAfterDuration = %s, want %s", d, maxRateLimitRetryDelay)
	}
}

func TestRetryAfterDuration_PastDate(t *testing.T) {
	retryAt := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
	d := retryAfterDuration(retryAt, time.Second)
	if d != 0 {
		t.Fatalf("retryAfterDuration for past date = %s, want 0", d)
	}
}

func TestRetryAfterDuration_InvalidHeader(t *testing.T) {
	d := retryAfterDuration("not-a-number", 3*time.Second)
	if d != 3*time.Second {
		t.Fatalf("retryAfterDuration = %s, want 3s fallback", d)
	}
}

func TestRetryAfterDuration_Empty(t *testing.T) {
	d := retryAfterDuration("", 2*time.Second)
	if d != 2*time.Second {
		t.Fatalf("retryAfterDuration = %s, want 2s fallback", d)
	}
}

func TestRetryAfterDuration_Whitespace(t *testing.T) {
	d := retryAfterDuration("  10  ", time.Second)
	if d != 10*time.Second {
		t.Fatalf("retryAfterDuration = %s, want 10s", d)
	}
}

// ========== retryDelay ==========

func TestRetryDelay_ZeroBase(t *testing.T) {
	if d := retryDelay(0, 5); d != 0 {
		t.Fatalf("retryDelay(0, 5) = %s, want 0", d)
	}
}

func TestRetryDelay_NegativeBase(t *testing.T) {
	if d := retryDelay(-1*time.Second, 3); d != 0 {
		t.Fatalf("retryDelay(-1s, 3) = %s, want 0", d)
	}
}

func TestRetryDelay_Attempt0(t *testing.T) {
	base := 100 * time.Millisecond
	if d := retryDelay(base, 0); d != base {
		t.Fatalf("retryDelay(100ms, 0) = %s, want 100ms", d)
	}
}

func TestRetryDelay_Attempt1(t *testing.T) {
	base := 100 * time.Millisecond
	if d := retryDelay(base, 1); d != 200*time.Millisecond {
		t.Fatalf("retryDelay(100ms, 1) = %s, want 200ms", d)
	}
}

func TestRetryDelay_Attempt2(t *testing.T) {
	base := 100 * time.Millisecond
	if d := retryDelay(base, 2); d != 400*time.Millisecond {
		t.Fatalf("retryDelay(100ms, 2) = %s, want 400ms", d)
	}
}

func TestRetryDelay_CappedAt30s(t *testing.T) {
	base := 10 * time.Second
	if d := retryDelay(base, 3); d != 30*time.Second {
		t.Fatalf("retryDelay(10s, 3) = %s, want 30s", d)
	}
}

func TestRetryDelay_CappedLoop(t *testing.T) {
	base := 1 * time.Second
	if d := retryDelay(base, 10); d != 30*time.Second {
		t.Fatalf("retryDelay(1s, 10) = %s, want 30s cap", d)
	}
}

// ========== waitForRetry ==========

func TestWaitForRetry_ZeroDelay(t *testing.T) {
	if err := waitForRetry(context.Background(), 0); err != nil {
		t.Fatalf("waitForRetry(0) err = %v", err)
	}
}

func TestWaitForRetry_NegativeDelay(t *testing.T) {
	if err := waitForRetry(context.Background(), -1*time.Second); err != nil {
		t.Fatalf("waitForRetry(-1s) err = %v", err)
	}
}

func TestWaitForRetry_CtxCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForRetry(ctx, 10*time.Second); err != context.Canceled {
		t.Fatalf("waitForRetry cancelled ctx err = %v, want context.Canceled", err)
	}
}

// ========== newQueryRateLimiter ==========

func TestNewQueryRateLimiter(t *testing.T) {
	l := newQueryRateLimiter(5 * time.Second)
	if l == nil || l.interval != 5*time.Second {
		t.Fatalf("newQueryRateLimiter interval = %s, want 5s", l.interval)
	}
}

// ========== withQuerySlot wait error path (needs shared path to avoid real lock) ==========

func TestWithQuerySlot_WaitErrorShared(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	// hold lock so reserve fails
	lock := flock.New(path + ".lock")
	locked, err := lock.TryLock()
	if err != nil {
		t.Fatalf("create lock: %v", err)
	}
	defer lock.Unlock()
	if !locked {
		t.Fatal("failed to acquire lock")
	}

	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err = limiter.withQuerySlot(ctx, func() error { return nil })
	if err == nil {
		t.Fatal("withQuerySlot should fail when shared lock is held")
	}
}

// ========== response with errmsg containing 45012 but error=false ==========

func TestIsRateLimitResponse_45012InErrMsgButErrorFalse(t *testing.T) {
	// error=false but errmsg contains 45012 - should NOT be rate limit
	if isRateLimitResponse(http.StatusOK, []byte(`{"error":false,"errmsg":"result id 45012"}`)) {
		t.Fatal("isRateLimitResponse should return false when error=false even if errmsg has 45012")
	}
}

// ========== reserveSharedSlotWithLock past timestamp ==========

func TestWithSharedQuerySlotAcceptsPastTimestamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	// write a state file with a timestamp in the past
	if err := os.WriteFile(path, []byte(strconv.FormatInt(time.Now().Add(-1*time.Hour).UnixNano(), 10)), 0o600); err != nil {
		t.Fatalf("write state: %v", err)
	}

	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	started := time.Now()
	called := false
	_, err := limiter.withQuerySlot(context.Background(), func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("request with past timestamp: %v", err)
	}
	if !called || time.Since(started) > 10*time.Millisecond {
		t.Fatalf("request called = %v and took %s; want immediate callback", called, time.Since(started))
	}
}

// ========== reserveSharedSlotWithLock write error ==========

func TestWithSharedQuerySlotReturnsStateReadError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("create state directory: %v", err)
	}

	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	called := false
	_, err := limiter.withQuerySlot(context.Background(), func() error {
		called = true
		return nil
	})
	if err == nil || called {
		t.Fatalf("request error = %v, called = %v; want state read error before callback", err, called)
	}
}

func TestReserveSharedSlotWithLockReturnsStateWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "query-rate-limit.state")

	limiter := newQueryRateLimiter(10 * time.Millisecond)
	limiter.sharedPath = path
	if _, err := limiter.reserveSharedSlotWithLock(); err == nil {
		t.Fatal("reserveSharedSlotWithLock should fail when the state directory does not exist")
	}
}

// ========== retryDelay cap after final doubling ==========

func TestRetryDelayCapsAfterFinalDoubling(t *testing.T) {
	// base=8s, attempt=2: 8->16->32, then cap the result at 30s.
	base := 8 * time.Second
	if d := retryDelay(base, 2); d != 30*time.Second {
		t.Fatalf("retryDelay(8s, 2) = %s, want 30s", d)
	}
}

// ========== withSharedQuerySlot waitForRetry error ==========

func TestWithSharedQuerySlot_WaitForRetryError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	limiter := newQueryRateLimiter(500 * time.Millisecond)
	limiter.sharedPath = path
	if _, err := limiter.withQuerySlot(context.Background(), func() error { return nil }); err != nil {
		t.Fatalf("first request: %v", err)
	}

	second := newQueryRateLimiter(500 * time.Millisecond)
	second.sharedPath = path
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(10*time.Millisecond, cancel)
	called := false
	_, err := second.withSharedQuerySlot(ctx, func() error {
		called = true
		return nil
	})
	if err != context.Canceled || called {
		t.Fatalf("cancelled shared request error = %v, called = %v; want context.Canceled before callback", err, called)
	}
}
