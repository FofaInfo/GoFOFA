package gofofa

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// params is key=>value for query, auto encoded with uri escape
func (c *Client) credentialValues() url.Values {
	values := url.Values{}
	if c.Email != "" {
		values.Set("email", c.Email)
	}
	values.Set("key", c.Key)
	return values
}

func (c *Client) buildURL(apiURI string, params map[string]string) string {
	fullURL := fmt.Sprintf("%s/api/%s/%s?", c.Server, c.APIVersion, apiURI)
	ps := c.credentialValues()
	for k, v := range params {
		ps.Set(k, v)
	}
	return fullURL + ps.Encode()
}

func readAll(reader io.Reader, size int) ([]byte, error) {
	if size <= 0 {
		size = int(math.Max(float64(size), 65535))
	}
	buffer := bytes.NewBuffer(make([]byte, 0, size))
	_, err := io.Copy(buffer, reader)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// just fetch fofa body, no need to unmarshal
func (c *Client) fetchBody(apiURI string, params map[string]string) (body []byte, err error) {
	ctx := c.GetContext()
	if ctx == nil {
		ctx = context.Background()
	}
	for attempt := 0; ; attempt++ {
		var statusCode int
		var retryAfter string
		if isQueryAPI(apiURI) {
			var waited time.Duration
			waited, err = c.queryLimiter.withQuerySlot(ctx, func() error {
				body, statusCode, retryAfter, err = c.fetchBodyOnce(ctx, apiURI, params)
				return err
			})
			if waited > 0 && c.logger != nil {
				c.logger.Debugf("fofa query rate-limit wait api=%s duration=%s", apiURI, waited)
			}
			if err != nil {
				return nil, err
			}
		} else {
			body, statusCode, retryAfter, err = c.fetchBodyOnce(ctx, apiURI, params)
		}
		if err != nil || !isRateLimitResponse(statusCode, body) || attempt >= c.rateLimitRetries {
			if err == nil && isRateLimitResponse(statusCode, body) && attempt >= c.rateLimitRetries {
				delay := retryAfterDuration(retryAfter, retryDelay(c.rateLimitRetryDelay, attempt))
				return nil, fmt.Errorf(
					"fofa rate limit response after %d retries (status=%d, retry-after=%s): %s",
					attempt, statusCode, delay, strings.TrimSpace(string(body)),
				)
			}
			return body, err
		}

		delay := retryAfterDuration(retryAfter, retryDelay(c.rateLimitRetryDelay, attempt))
		if c.logger != nil {
			c.logger.Warnf("fofa rate-limit retry api=%s attempt=%d delay=%s", apiURI, attempt+1, delay)
		}
		if err = waitForRetry(ctx, delay); err != nil {
			return nil, err
		}
	}
}

func (c *Client) fetchBodyOnce(ctx context.Context, apiURI string, params map[string]string) (body []byte, statusCode int, retryAfter string, err error) {
	var req *http.Request
	var resp *http.Response

	fullURL := c.buildURL(apiURI, params)
	if c.logger != nil {
		c.logger.Debugf("fetch fofa: %s", apiURI)
	}
	//c.logger.Debugf("fetch fofa: %s", fullURL)

	req, err = http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Accept-Encoding", "gzip")
	//requestDump, _ := httputil.DumpRequestOut(req, false)
	//log.Println(string(requestDump))

	resp, err = c.httpClient.Do(req)
	//responseDump, _ := httputil.DumpResponse(resp, false)
	//log.Println(string(responseDump))
	if err != nil {
		if ctx.Err() != nil {
			return nil, 0, "", ctx.Err()
		}
		if !c.accountDebug {
			// 替换账号明文信息
			if e, ok := err.(*url.Error); ok {
				redactedClient := *c
				if c.Email != "" {
					redactedClient.Email = "<email>"
				}
				redactedClient.Key = "<key>"
				e.URL = redactedClient.buildURL(apiURI, params)
				err = e
			}
		}
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	retryAfter = resp.Header.Get("Retry-After")

	contentLength := 0
	if v := resp.Header.Get("Content-Length"); len(v) > 0 {
		contentLength, err = strconv.Atoi(v)
		if err != nil {
			return nil, statusCode, retryAfter, fmt.Errorf("invalid Content-Length %q: %w", v, err)
		}
	}
	encoding := resp.Header.Get("Content-Encoding")
	// 取body
	body, err = readAll(resp.Body, contentLength)
	if err != nil {
		return nil, statusCode, retryAfter, err
	}

	switch encoding {
	case "gzip":
		var reader *gzip.Reader
		reader, err = gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, statusCode, retryAfter, err
		}
		body, err = readAll(reader, 0)
		if err != nil {
			return nil, statusCode, retryAfter, err
		}
		if err = reader.Close(); err != nil {
			return nil, statusCode, retryAfter, err
		}
		//case "deflate":
		//	reader1 := flate.NewReader(bytes.NewReader(body))
		//	body, err = readAll(reader1, 0)
		//	if err != nil {
		//		return
		//	}
		//	reader1.Close()
	}
	if encoding == "gzip" {

	}

	//respDump, _ := httputil.DumpResponse(resp, false)
	//logrus.Debugln(string(respDump))

	return body, statusCode, retryAfter, nil
}

// Fetch http request and parse as json return to v
func (c *Client) Fetch(apiURI string, params map[string]string, v interface{}) (err error) {
	content, err := c.fetchBody(apiURI, params)
	if err != nil {
		return
	}

	if err = json.Unmarshal(content, v); err != nil {
		return
	}
	return
}
