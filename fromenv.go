package gofofa

import (
	"context"
	"os"
	"time"
)

// env can set:FOFA_SERVER,FOFA_KEY,FOFA_CLIENT_URL; FOFA_EMAIL is optional for legacy configurations.
// FOFA_CLIENT_URL > FOFA_SERVER
func newClientFromEnv() (*Client, error) {
	c := &Client{
		Server:              defaultServer,
		APIVersion:          defaultAPIVersion,
		ctx:                 context.Background(),
		rateLimitRetries:    3,
		rateLimitRetryDelay: time.Second,
	}

	if v := os.Getenv("FOFA_SERVER"); len(v) > 0 {
		c.Server = v
	}
	if v := os.Getenv("FOFA_EMAIL"); len(v) > 0 {
		c.Email = v
	}
	if v := os.Getenv("FOFA_KEY"); len(v) > 0 {
		c.Key = v
	}
	if v := os.Getenv("FOFA_CLIENT_URL"); len(v) > 0 {
		if err := c.Update(v); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// FofaURLFromEnv parse fofa connection url from env, then generate url string
func FofaURLFromEnv() string {
	c, err := newClientFromEnv()
	if err != nil {
		return ""
	}

	return c.URL()
}
