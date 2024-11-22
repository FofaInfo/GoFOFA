/*
Package gofofa fofa client in Go

env settings:
- FOFA_CLIENT_URL full fofa connnection string, format: <url>/?email=<email>&key=<key>&version=<v2>
- FOFA_SERVER fofa server
- FOFA_EMAIL fofa account email
- FOFA_KEY fofa account key
*/
package gofofa

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
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

	onResults    func(results [][]string) // when fetch results callback
	accountDebug bool                     // 调试账号明文信息
}

// Update merge config from config url
func (c *Client) Update(configURL string) error {
	u, err := url.Parse(configURL)
	if err != nil {
		return err
	}

	c.Server = u.Scheme + "://" + u.Host
	if u.Query().Has("email") {
		c.Email = u.Query().Get("email")
	}

	if u.Query().Has("key") {
		c.Key = u.Query().Get("key")
	}

	if u.Query().Has("version") {
		c.APIVersion = u.Query().Get("version")
	}

	return nil
}

// URL generate fofa connection url string
func (c *Client) URL() string {
	return fmt.Sprintf("%s/?email=%s&key=%s&version=%s", c.Server, c.Email, c.Key, c.APIVersion)
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

// WithURL configURL format: <url>/?email=<email>&key=<key>&version=<v2>&tlsdisabled=false&debuglevel=0
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
		return c, err
	}

	if c.Account.Error {
		c.logger.Warnf("auth failed")
		return c, fmt.Errorf("auth failed: '%s', make sure key is valid", c.Account.ErrMsg)
	}

	return c, nil
}
