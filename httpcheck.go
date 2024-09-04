package gofofa

import (
	"errors"
	"git.gobies.org/goby/httpclient"
	"net"
	"strconv"
)

type HttpResponse struct {
	IsActive   bool
	StatusCode string
}

func DoHttpCheck(rowURL string, retry int) HttpResponse {
	fURL := httpclient.NewFixUrl(rowURL)
	cfg := httpclient.NewGetRequestConfig("/")
	resp, err := retryDoHttpRequest(fURL, cfg, retry)
	if err != nil {
		return HttpResponse{false, "0"}
	}

	return HttpResponse{true, strconv.Itoa(resp.StatusCode)}
}

func retryDoHttpRequest(hostinfo *httpclient.FixUrl, req *httpclient.RequestConfig, retry int) (*httpclient.HttpResponse, error) {
	for i := 0; i < retry; i++ {
		resp, err := httpclient.DoHttpRequest(hostinfo, req)
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
