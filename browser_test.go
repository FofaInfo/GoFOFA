package gofofa

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func runningTime() func() {
	starTime := time.Now()
	return func() {
		fmt.Println("运行耗时 ==:", time.Now().Sub(starTime))
	}
}

func TestWorkerBrowser_Run(t *testing.T) {
	defer runningTime()()

	redirectTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html lang="en">
			<head><title>Successfully Title</title></head>
			<body>
				<h1>Redirected Successfully!</h1>
			</body>
			</html>
		`)
	}))
	defer redirectTarget.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/js/normal":
			fmt.Fprint(w, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Initial Title</title>
				<script>
					// 在页面加载后 1 秒更新标题
					setTimeout(function() {
						document.title = "Updated Title";
					}, 1000);
				</script>
			</head>
			<body>
				<p>Waiting for title update...</p>
			</body>
			</html>
			`)
			return
		case "/js/redirect":
			fmt.Fprintf(w, `
				<!DOCTYPE html>
				<html lang="en">
				<head>
					<meta charset="UTF-8">
					<title>Redirect Page</title>
					<script>
						// 页面加载后 1 秒跳转到目标页面
						setTimeout(function() {
							window.location.href = "%s/target";
						}, 1000);
					</script>
				</head>
				<body>
					<p>Redirecting...</p>
				</body>
				</html>
			`, redirectTarget.URL)
			return
		}
	}))
	defer ts.Close()

	// 错误url情况
	b := NewWorkerBrowser("")
	body, err := b.Run()
	assert.NotNil(t, err)
	assert.Nil(t, body["body"])

	// 常规js渲染
	b = NewWorkerBrowser(ts.URL + "/js/normal")
	body, err = b.Run()
	assert.Nil(t, err)
	assert.Equal(t, "Updated Title", body["title"])

	// 页面跳转
	b = NewWorkerBrowser(ts.URL + "/js/redirect")
	body, err = b.Run()
	assert.Nil(t, err)
	assert.Equal(t, "Successfully Title", body["title"])
}
