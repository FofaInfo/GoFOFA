package gofofa

import (
	"errors"
	"git.gobies.org/goby/httpclient"
	"net"
	"strconv"
)

type Result struct {
	IsActive   bool
	StatusCode string
}

func DoHttpCheck(rowURL string, retry int) Result {
	fURL := httpclient.NewFixUrl(rowURL)
	if fURL == nil {
		return Result{false, "0"}
	}
	cfg := httpclient.NewGetRequestConfig("/")
	resp, err := retryDoHttpRequest(fURL, cfg, retry)
	if err != nil {
		return Result{false, "0"}
	}

	return Result{true, strconv.Itoa(resp.StatusCode)}
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
