package gofofa

import "git.gobies.org/goby/httpclient"

func CheckActive(fixedHostInfo string) bool {
	fURL := httpclient.NewFixUrl(fixedHostInfo)
	cfg := httpclient.NewGetRequestConfig("/")
	_, err := httpclient.DoHttpRequest(fURL, cfg)
	if err != nil {
		return false
	}
	return true
}
