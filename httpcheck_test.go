package gofofa

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckActive(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer httpServer.Close()

	httpsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer httpsServer.Close()

	notFoundServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer notFoundServer.Close()

	closedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closedServer.URL
	closedServer.Close()

	tests := []struct {
		name          string
		fixedHostInfo string
		want          HttpResponse
	}{
		{
			name:          "HTTP URL",
			fixedHostInfo: httpServer.URL,
			want:          HttpResponse{IsActive: true, StatusCode: "200"},
		},
		{
			name:          "Connection failure",
			fixedHostInfo: closedURL,
			want:          HttpResponse{IsActive: false, StatusCode: "0"},
		},
		{
			name:          "IP and port without scheme",
			fixedHostInfo: strings.TrimPrefix(httpServer.URL, "http://"),
			want:          HttpResponse{IsActive: true, StatusCode: "200"},
		},
		{
			name:          "HTTP error status is active",
			fixedHostInfo: notFoundServer.URL,
			want:          HttpResponse{IsActive: true, StatusCode: "404"},
		},
		{
			name:          "HTTPS URL",
			fixedHostInfo: httpsServer.URL,
			want:          HttpResponse{IsActive: true, StatusCode: "200"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DoHttpCheck(tt.fixedHostInfo, 3), "CheckActive(%v)", tt.fixedHostInfo)
		})
	}

}
