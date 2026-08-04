package gofofa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

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

func TestHostSearchDoesNotRetrySuccessful45012Result(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/all" {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{"error": true})
			return
		}
		calls++
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"error":   false,
			"results": [][]string{{"45012"}},
		})
	}))
	defer server.Close()

	client := newRateLimitTestClient(server)
	results, err := client.HostSearch("port=80", 1, []string{"ip"})
	if err != nil {
		t.Fatalf("HostSearch returned error: %v", err)
	}
	if calls != 1 || len(results) != 1 || results[0][0] != "45012" {
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
	if callbackCalls != 2 {
		t.Fatalf("callback count = %d, want 2 delivered pages", callbackCalls)
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
	if calls != 2 || callbackCalls != 2 {
		t.Fatalf("calls=%d callbacks=%d, want two delivered pages", calls, callbackCalls)
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

	if waited, err := first.wait(context.Background()); err != nil || waited > 5*time.Millisecond {
		t.Fatalf("first wait = %s, err=%v; want immediate reservation", waited, err)
	}
	if waited, err := second.wait(context.Background()); err != nil {
		t.Fatalf("second wait returned error: %v", err)
	} else if waited < 15*time.Millisecond {
		t.Fatalf("second wait = %s, want shared reservation delay", waited)
	}
}

func TestSharedQueryRateLimiterRejectsCorruptState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	if err := os.WriteFile(path, []byte("corrupt"), 0o600); err != nil {
		t.Fatalf("write corrupt state: %v", err)
	}
	limiter := newQueryRateLimiter(20 * time.Millisecond)
	limiter.sharedPath = path
	if _, err := limiter.wait(context.Background()); err == nil || !strings.Contains(err.Error(), "invalid shared rate limit state") {
		t.Fatalf("wait error = %v, want corrupt state error", err)
	}
}

func TestSharedQueryRateLimiterRejectsEmptyState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "query-rate-limit.state")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("write empty state: %v", err)
	}
	limiter := newQueryRateLimiter(20 * time.Millisecond)
	limiter.sharedPath = path
	if _, err := limiter.wait(context.Background()); err == nil || !strings.Contains(err.Error(), "invalid shared rate limit state") {
		t.Fatalf("wait error = %v, want empty state error", err)
	}
}
