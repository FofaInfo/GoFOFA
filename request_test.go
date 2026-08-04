package gofofa

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var (
	fetchHander = func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/gzip.json":
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			gz.Write([]byte(`{"text":"hello world"}`))
			return
		case "/api/v1/contentLengthError.json":
			w.Header().Set("Content-Length", "aaa")
			return
		}
	}
)

type tcpTestServer struct {
	listener net.Listener
}

func (ts *tcpTestServer) URL() string {
	return "http://" + ts.listener.Addr().String()
}

func (ts *tcpTestServer) Close() {
	ts.listener.Close()
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

func newTcpTestServer(handler func(conn net.Conn, data []byte) error) *tcpTestServer {
	ts := &tcpTestServer{
		listener: newLocalListener(),
	}

	go func() {
		data := make([]byte, 1024)
		var err error
		var n int
		var conn net.Conn

		for {
			conn, err = ts.listener.Accept()
			if err != nil {
				break
			}

			n, err = conn.Read(data)
			if err != nil {
				break
			}

			err = handler(conn, data[:n])
			if err != nil {
				break
			}
			conn.Close()
		}
	}()

	return ts
}

func TestClient_Fetch(t *testing.T) {
	_, err := NewClient(WithURL("http://127.0.0.1:55"))
	assert.Error(t, err)

	ts := httptest.NewServer(http.HandlerFunc(fetchHander))
	defer ts.Close()

	cli := &Client{
		Server:     ts.URL,
		APIVersion: "v1",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}

	// 解析异常
	var a map[string]interface{}
	err = cli.Fetch("", nil, &a)
	assert.Error(t, err)

	// gzip
	err = cli.Fetch("gzip.json", nil, &a)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", a["text"].(string))

	// content Length Error
	err = cli.Fetch("contentLengthError.json", nil, &a)
	assert.Contains(t, err.Error(), "unexpected end of JSON input")

	// read all error 需要构造一个恶意服务器
	s := newTcpTestServer(func(conn net.Conn, data []byte) error {
		_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 10\r\nConnection: close\r\n\r\na"))
		return err
	})
	defer s.Close()
	cli = &Client{
		Server:     s.URL(),
		APIVersion: "v1",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	err = cli.Fetch("/", nil, &a)
	assert.Contains(t, err.Error(), "unexpected EOF")

	// 构造content length 的atoi 异常， 需要构造一个恶意服务器
	s1 := newTcpTestServer(func(conn net.Conn, data []byte) error {
		_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: -\r\nConnection: close\r\n\r\na"))
		return err
	})
	defer s1.Close()
	cli = &Client{
		Server:     s1.URL(),
		APIVersion: "v1",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	err = cli.Fetch("/", nil, &a)
	assert.Contains(t, err.Error(), "bad Content-Length")

	// 构造错误的gzip， 需要构造一个恶意服务器
	s2 := newTcpTestServer(func(conn net.Conn, data []byte) error {
		_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 1\r\nContent-Encoding: gzip\r\nConnection: close\r\n\r\na"))
		return err
	})
	defer s2.Close()
	cli = &Client{
		Server:     s2.URL(),
		APIVersion: "v1",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	err = cli.Fetch("/", nil, &a)
	assert.Contains(t, err.Error(), "unexpected EOF")
}

func TestFetchNetworkErrorDoesNotMutateCredentials(t *testing.T) {
	calls := 0
	var secondRequest *http.Request
	cli := &Client{
		Server:     "http://fofa.test",
		APIVersion: "v1",
		Email:      "account@example.com",
		Key:        "secret-key",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	cli.httpClient.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("network failure")
		}
		secondRequest = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"error":false}`)),
			Request:    req,
		}, nil
	})

	var result HostResults
	if err := cli.Fetch("search/all", nil, &result); err == nil {
		t.Fatal("first request error = nil, want network error")
	}
	if err := cli.Fetch("search/all", nil, &result); err != nil {
		t.Fatalf("second request error = %v", err)
	}
	if cli.Email != "account@example.com" || cli.Key != "secret-key" {
		t.Fatalf("client credentials changed: email=%q key=%q", cli.Email, cli.Key)
	}
	if secondRequest.URL.Query().Get("email") != "account@example.com" || secondRequest.URL.Query().Get("key") != "secret-key" {
		t.Fatalf("second request credentials changed: %s", secondRequest.URL)
	}
}

func TestBuildURLOmitsEmptyEmail(t *testing.T) {
	client := &Client{
		Server:     "https://fofa.info",
		APIVersion: "v1",
		Key:        "secret-key",
	}

	requestURL, err := url.Parse(client.buildURL("search/all", nil))
	if err != nil {
		t.Fatalf("parse request URL: %v", err)
	}
	query := requestURL.Query()
	if query.Has("email") {
		t.Fatalf("request URL contains empty email: %s", requestURL)
	}
	if query.Get("key") != "secret-key" {
		t.Fatalf("request URL key = %q, want secret-key", query.Get("key"))
	}
}
