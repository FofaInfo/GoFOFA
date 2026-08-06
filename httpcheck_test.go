package gofofa

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoHttpCheck(t *testing.T) {
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer successServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer failServer.Close()

	tests := []struct {
		name string
		url  string
		want HttpResponse
	}{
		{
			name: "Success",
			url:  successServer.URL,
			want: HttpResponse{IsActive: true, StatusCode: "200"},
		},
		{
			name: "404",
			url:  failServer.URL,
			want: HttpResponse{IsActive: true, StatusCode: "404"},
		},
		{
			name: "Invalid host",
			url:  "http://127.0.0.1:1",
			want: HttpResponse{IsActive: false, StatusCode: "0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DoHttpCheck(tt.url, 1)
			assert.Equal(t, tt.want, got)
		})
	}
}
