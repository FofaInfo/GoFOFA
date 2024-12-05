package gofofa

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"golang.org/x/net/html"
	"log"
	"strings"
	"time"
)

type WorkerBrowser struct {
	Url string
}

func NewWorkerBrowser(url string) *WorkerBrowser {
	return &WorkerBrowser{
		Url: url,
	}
}

func (wp *WorkerBrowser) Run() (response map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred: %v", r)
			log.Println("browser render error:", r, "Url:", wp.Url)
		}
	}()

	if wp.Url == "" {
		return nil, errors.New("url is empty")
	}
	body, err := wp.renderScan(wp.Url)
	if err != nil {
		return nil, errors.New("browser render error" + err.Error())
	}

	response = make(map[string]interface{})
	response["url"] = wp.Url

	if !strings.Contains(body, "<title>") {
		response["body"] = body
		return response, nil
	}

	title := strings.TrimSpace(wp.ParseHTML(body, "title"))
	title = wp.removeExtraSpaces(title)

	response["body"] = body
	response["title"] = title
	return response, nil
}

func (wp *WorkerBrowser) ParseHTML(htmlStr string, tag string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		log.Println("parsing HTML error", err)
		return ""
	}
	return wp.findNodeText(doc, tag)
}

func (wp *WorkerBrowser) removeExtraSpaces(input string) string {
	// 将空白字符都转换成空格
	input = strings.ReplaceAll(input, "\n", ` `)
	input = strings.ReplaceAll(input, "\t", ` `)
	input = strings.ReplaceAll(input, "\r", ` `)

	var builder strings.Builder
	wasSpace := false

	// 遍历输入字符串的每个字符
	for _, char := range input {
		if char == ' ' {
			if wasSpace {
				continue
			}
			wasSpace = true
		} else {
			wasSpace = false
		}
		builder.WriteRune(char)
	}

	return builder.String()
}

func (wp *WorkerBrowser) findNodeText(n *html.Node, tag string) string {
	if n.Type == html.ElementNode && n.Data == tag {
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			return n.FirstChild.Data
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := wp.findNodeText(c, tag)
		if result != "" {
			return result
		}
	}
	return ""
}

func (wp *WorkerBrowser) renderScan(url string) (string, error) {
	const maxRetries = 3
	const retryDelay = 5 * time.Second
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		bodyHTML, err := func() (string, error) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			dev := devices.Device{
				Title:          "Laptop with MDPI screen",
				Capabilities:   []string{},
				UserAgent:      `Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`,
				AcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
				Screen: devices.Screen{
					DevicePixelRatio: 1,
					Horizontal: devices.ScreenSize{
						Width:  1280,
						Height: 800,
					},
					Vertical: devices.ScreenSize{
						Width:  800,
						Height: 1280,
					},
				},
			}

			l := launcher.New().
				Headless(true).
				Devtools(true)
			defer l.Cleanup()

			l.Set("disable-web-security")
			l.Set("allow-running-insecure-content")
			l.Set("--ignore-certificate-errors")
			l.Set("disable-notifications", "true")

			lurl := l.MustLaunch()

			b := rod.New().ControlURL(lurl).
				DefaultDevice(dev).
				Timeout(10 * time.Second).
				Trace(false).
				MustConnect()
			defer b.MustClose()

			page := b.MustPage().Context(ctx)
			defer page.MustClose()

			// 启动协程监听浏览器弹窗事件
			go func() {
				page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
					_ = proto.PageHandleJavaScriptDialog{Accept: false, PromptText: ""}.Call(page)
				})()

				select {
				case <-ctx.Done():
					return
				}
			}()

			if err := page.Navigate(url); err != nil {
				return "", fmt.Errorf("navigation error: %v", err)
			}

			if err := page.WaitLoad(); err != nil {
				return "", fmt.Errorf("page load error: %v", err)
			}

			time.Sleep(3 * time.Second)

			// 获取页面的 HTML 内容
			body, err := page.Element("html")
			if err != nil {
				return "", fmt.Errorf("not find html element: %v", err)
			}

			bodyHtml, err := body.HTML()
			if err != nil {
				return "", fmt.Errorf("failed to get html content: %v", err)
			}

			return bodyHtml, nil
		}()

		if err == nil && bodyHTML != "" {
			return bodyHTML, nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v", attempt, lastErr)
		time.Sleep(retryDelay)

	}

	return "", fmt.Errorf("all attempts failed: %v", lastErr)
}
