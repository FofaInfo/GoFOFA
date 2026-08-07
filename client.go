/*
Package gofofa fofa client in Go

env settings:
- FOFA_CLIENT_URL full fofa connection string, format: <url>/?key=<key>&version=<v2>
- FOFA_SERVER fofa server
- FOFA_EMAIL optional legacy account email
- FOFA_KEY fofa account key
*/
package gofofa

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultServer     = "https://fofa.info"
	defaultAPIVersion = "v1"
)

// Client of fofa connection
type Client struct {
	Server     string // can set local server for debugging, format: <scheme>://<host>
	APIVersion string // api version
	Email      string // fofa email
	Key        string // fofa key

	Account    AccountInfo // fofa account info
	DeductMode DeductMode  // Deduct Mode

	httpClient *http.Client //
	logger     *logrus.Logger
	ctx        context.Context // use to cancel requests

	onResults           func(results [][]string) // when fetch results callback
	accountDebug        bool                     // 调试账号明文信息
	queryLimiter        *queryRateLimiter
	rateLimitRetries    int
	rateLimitRetryDelay time.Duration
	sharedRateLimit     bool
}

// Update merge config from config url
func (c *Client) Update(configURL string) error {
	u, err := url.Parse(configURL)
	if err != nil {
		return err
	}

	c.Server = u.Scheme + "://" + u.Host
	query := u.Query()
	if query.Has("email") {
		c.Email = query.Get("email")
	}

	if query.Has("key") {
		c.Key = query.Get("key")
		if !query.Has("email") {
			c.Email = ""
		}
	}

	if query.Has("version") {
		c.APIVersion = query.Get("version")
	}

	return nil
}

// URL generate fofa connection url string
func (c *Client) URL() string {
	params := url.Values{}
	if c.Email != "" {
		params.Set("email", c.Email)
	}
	params.Set("key", c.Key)
	params.Set("version", c.APIVersion)
	return fmt.Sprintf("%s/?%s", c.Server, params.Encode())
}

// GetContext 获取context，用于中止任务
func (c *Client) GetContext() context.Context {
	return c.ctx
}

// SetContext 设置context，用于中止任务
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

type ClientOption func(c *Client) error

// WithURL configURL format: <url>/?key=<key>&version=<v2>&tlsdisabled=false&debuglevel=0
// The legacy email parameter is optional.
func WithURL(configURL string) ClientOption {
	return func(c *Client) error {
		// merge from config
		if len(configURL) > 0 {
			return c.Update(configURL)
		}
		return nil
	}
}

// WithLogger set logger
func WithLogger(logger *logrus.Logger) ClientOption {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

// WithOnResults set on results callback
func WithOnResults(onResults func(results [][]string)) ClientOption {
	return func(c *Client) error {
		c.onResults = onResults
		return nil
	}
}

// WithAccountDebug 是否错误里面显示账号密码原始信息
func WithAccountDebug(v bool) ClientOption {
	return func(c *Client) error {
		c.accountDebug = v
		return nil
	}
}

// WithRateLimitRetry configures retries for FOFA rate-limit responses.
// maxRetries is the number of attempts after the initial request.
func WithRateLimitRetry(maxRetries int, retryDelay time.Duration) ClientOption {
	return func(c *Client) error {
		if maxRetries < 0 {
			return fmt.Errorf("maxRetries must be non-negative")
		}
		if retryDelay < 0 {
			return fmt.Errorf("retryDelay must be non-negative")
		}
		c.rateLimitRetries = maxRetries
		c.rateLimitRetryDelay = retryDelay
		return nil
	}
}

// WithSharedRateLimit enables a process-shared query rate limit for CLI-style
// clients. The state file contains only the next send time, not credentials.
func WithSharedRateLimit(enabled bool) ClientOption {
	return func(c *Client) error {
		c.sharedRateLimit = enabled
		return nil
	}
}

// NewClient from fofa connection string to config
// and with env config merge
func NewClient(options ...ClientOption) (*Client, error) {
	// read from env
	c, err := newClientFromEnv()
	if err != nil {
		return c, err
	}

	c.logger = logrus.New()
	for _, opt := range options {
		err = opt(c)
		if err != nil {
			return nil, err
		}
	}

	// fetch one time to make sure network is ok
	c.httpClient = &http.Client{}
	c.Account, err = c.AccountInfo()
	if err != nil {
		c.logger.Warnf("account invalid")
		if c.Account.Error {
			c.logger.Warnf("auth failed")
			message := c.Account.ErrMsg
			if message == "" {
				message = err.Error()
			}
			return c, fmt.Errorf("auth failed: '%s', make sure key is valid", message)
		}
		return c, err
	}
	c.queryLimiter = newQueryRateLimiter(queryInterval(c.Account))
	if c.sharedRateLimit && c.queryLimiter.interval > 0 {
		c.queryLimiter.sharedPath = sharedRateLimitPath(c.Server, c.Key)
	}

	return c, nil
}
