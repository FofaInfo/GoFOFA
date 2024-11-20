package gofofa

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HttpResponse struct {
	IsActive   bool
	StatusCode string
}

func NewFixUrl(rowURL string) string {
	fullURL := rowURL
	hasScheme := strings.Contains(rowURL, "://")
	if !hasScheme {
		fullURL = "http://" + fullURL
	}
	fullURL = strings.Trim(fullURL, " \t\r\n")
	return fullURL
}

func NewRequestConfig(fullURL string) *http.Client {
	client := &http.Client{
		Timeout: time.Second * time.Duration(30), // 超时时间
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 禁止自动跳转
		},
	}
	if strings.HasPrefix(fullURL, "https://") {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS10},
		}
	}
	return client
}

func DoHttpCheck(rowURL string, retry int) HttpResponse {
	log.Println("check active of:", rowURL)
	fURL := NewFixUrl(rowURL)
	client := NewRequestConfig(fURL)
	resp, err := retryDoHttpRequest(client, fURL, retry)
	if err != nil {
		log.Println("check active of:", rowURL, "error:", err)
		return HttpResponse{false, "0"}
	}

	return HttpResponse{true, strconv.Itoa(resp.StatusCode)}
}

func retryDoHttpRequest(client *http.Client, url string, retry int) (*http.Response, error) {
	for i := 0; i < retry; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			var netError net.Error
			if errors.As(err, &netError) {
				if netError.Timeout() {
					continue
				}
			}
			return nil, err
		}
		return resp, nil
	}
	return nil, errors.New("retry exceeded")
}
