package gofofa

import (
	"git.gobies.org/goby/httpclient"
	"strconv"
)

func CheckActive(fixedHostInfo string) bool {
	fURL := httpclient.NewFixUrl(fixedHostInfo)
	cfg := httpclient.NewGetRequestConfig("/")
	_, err := httpclient.DoHttpRequest(fURL, cfg)
	if err != nil {
		return false
	}
	return true
}

func HandleStatusCode(fixedHostInfo string) string {
	fURL := httpclient.NewFixUrl(fixedHostInfo)
	cfg := httpclient.NewGetRequestConfig("/")
	resp, err := httpclient.DoHttpRequest(fURL, cfg)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(resp.StatusCode)
}
