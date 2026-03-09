package cmd

import (
	stdjson "encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	gofofa "github.com/FofaInfo/GoFOFA"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

type HostResults struct {
	Error   bool          `json:"error"`
	Size    int           `json:"size"`
	Results []interface{} `json:"results"`
}

func Test_constructQuery(t *testing.T) {
	tests := []struct {
		name      string
		queryType string
		queries   []string
		want      string
	}{
		{
			name:      "single query",
			queryType: "ip",
			queries:   []string{"1.1.1.1"},
			want:      "ip=1.1.1.1",
		},
		{
			name:      "multiple queries",
			queryType: "domain",
			queries:   []string{"google.com", "baidu.com"},
			want:      "domain=google.com || domain=baidu.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := constructQuery(tt.queryType, tt.queries); got != tt.want {
				t.Errorf("constructQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_batchProcess(t *testing.T) {
	tests := []struct {
		name      string
		queries   []string
		batchSize int
		queryType string
		want      []string
	}{
		{
			name:      "batch size 1",
			queries:   []string{"1.1.1.1", "2.2.2.2"},
			batchSize: 1,
			queryType: "ip",
			want:      []string{"ip=1.1.1.1", "ip=2.2.2.2"},
		},
		{
			name:      "batch size 2",
			queries:   []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"},
			batchSize: 2,
			queryType: "ip",
			want:      []string{"ip=1.1.1.1 || ip=2.2.2.2", "ip=3.3.3.3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := batchProcess(tt.queries, tt.batchSize, tt.queryType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("batchProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Mocking gofofa.Client is hard because it's a struct and many methods are on it.
// However, we can mock the network if we want, or in this case, we just want to test
// that DumpAction calls DumpSearch correctly.
// Since fofaCli is global, we can swap it.

type mockDumpClient struct {
	*gofofa.Client
	dumpSearchFunc func(query string, size int, batchSize int, fields []string, callback func([][]string, int) error, options gofofa.SearchOptions) error
}

func (m *mockDumpClient) DumpSearch(query string, size int, batchSize int, fields []string, callback func([][]string, int) error, options gofofa.SearchOptions) error {
	return m.dumpSearchFunc(query, size, batchSize, fields, callback, options)
}

// Note: DumpSearch is not an interface method, so mockDumpClient doesn't actually override gofofa.Client's method
// when called via a gofofa.Client pointer. We'd need an interface for fofaCli.
// For now, let's at least test the helper functions.
// If we want to test DumpAction, we might need to refactor gofofa.Client to an interface.
// But the user just asked for "corresponding unit tests".

func TestDumpAction_Concurrency(t *testing.T) {
	// 1. Setup mock server
	var mu sync.Mutex
	queryCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		queryCount++
		mu.Unlock()

		if strings.Contains(r.URL.Path, "/api/v1/info/my") {
			w.Write([]byte(`{"error":false, "fcoin":100, "vip_level":2}`))
			return
		}

		if strings.Contains(r.URL.Path, "search/next") {
			// Simulate FOFA response
			res := HostResults{
				Error: false,
				Size:  1,
				Results: []interface{}{
					[]interface{}{"1.1.1.1", "80"},
				},
			}
			data, _ := stdjson.Marshal(res)
			w.Write(data)
			return
		}
	}))
	defer ts.Close()

	// 2. Mock fofaCli
	origCli := fofaCli
	defer func() { fofaCli = origCli }()

	var err error
	fofaCli, err = gofofa.NewClient(gofofa.WithURL(ts.URL + "/?email=test&key=test"))
	assert.Nil(t, err)

	// 3. Setup CLI App
	app := &cli.App{
		Commands: []*cli.Command{
			dumpCmd,
		},
	}

	// 4. Create inFile with multiple queries
	tmpFile, _ := os.CreateTemp("", "queries.txt")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("domain=google.com\ndomain=baidu.com\ndomain=bing.com\n")
	tmpFile.Close()

	// 5. Run DumpAction with concurrency
	// We need to capture stdout or just check if it doesn't fail
	err = app.Run([]string{"fofa", "dump", "-i", tmpFile.Name(), "-workers", "3", "-s", "1"})
	assert.Nil(t, err)

	// 6. Verify all queries were processed
	assert.Equal(t, 3, queryCount-1) // -1 for the initial info/my call in NewClient
}

func TestDumpAction_DefaultsWithInFile(t *testing.T) {
	// 1. Setup mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/info/my") {
			w.Write([]byte(`{"error":false, "fcoin":100, "vip_level":2}`))
			return
		}
		if strings.Contains(r.URL.Path, "search/next") {
			// Simulate FOFA response
			res := HostResults{
				Error: false,
				Size:  1,
				Results: []interface{}{
					[]interface{}{"1.1.1.1", "80"},
				},
			}
			data, _ := stdjson.Marshal(res)
			w.Write(data)
			return
		}
	}))
	defer ts.Close()

	// 2. Mock fofaCli
	origCli := fofaCli
	defer func() { fofaCli = origCli }()

	var err error
	fofaCli, err = gofofa.NewClient(gofofa.WithURL(ts.URL + "/?email=test&key=test"))
	assert.Nil(t, err)

	// 3. Setup CLI App
	app := &cli.App{
		Commands: []*cli.Command{
			dumpCmd,
		},
	}

	// 4. Create inFile with multiple queries
	tmpFile, _ := os.CreateTemp("", "queries.txt")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("domain=google.com\n")
	tmpFile.Close()

	// reset globals
	workers = 1
	ratePerSecond = 2 // default is 2 anyway, but let's change it to test
	origRate := ratePerSecond
	ratePerSecond = 1
	defer func() { ratePerSecond = origRate }()

	// 5. Run DumpAction without workers and rate explicit flags
	err = app.Run([]string{"fofa", "dump", "-i", tmpFile.Name(), "-s", "1"})
	assert.Nil(t, err)

	// 6. Verify workers and rate are updated to 10 and 2
	assert.Equal(t, 10, workers)
	assert.Equal(t, 2, ratePerSecond)
}
