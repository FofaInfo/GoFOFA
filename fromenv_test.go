package gofofa

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFofaURLFromEnv(t *testing.T) {
	defer func() {
		os.Unsetenv("FOFA_CLIENT_URL")
		os.Unsetenv("FOFA_SERVER")
		os.Unsetenv("FOFA_EMAIL")
		os.Unsetenv("FOFA_KEY")
	}()

	os.Setenv("FOFA_SERVER", "https://1.1.1.1")
	os.Setenv("FOFA_KEY", "123456")
	os.Unsetenv("FOFA_CLIENT_URL")
	assert.Equal(t, "https://1.1.1.1/?key=123456&version=v1", FofaURLFromEnv())

	// Legacy email configuration remains supported.
	os.Setenv("FOFA_EMAIL", "a@a.com")
	assert.Equal(t, "https://1.1.1.1/?email=a%40a.com&key=123456&version=v1", FofaURLFromEnv())

	// 异常
	os.Setenv("FOFA_CLIENT_URL", "\x7f")
	assert.Equal(t, "", FofaURLFromEnv())

	// 部分更新
	os.Setenv("FOFA_CLIENT_URL", "https://2.2.2.2/?email=b@b.com")
	assert.Equal(t, "https://2.2.2.2/?email=b%40b.com&key=123456&version=v1", FofaURLFromEnv())

	// 全更新
	os.Setenv("FOFA_CLIENT_URL", "https://2.2.2.2/?email=b@b.com&key=000000&version=v2")
	assert.Equal(t, "https://2.2.2.2/?email=b%40b.com&key=000000&version=v2", FofaURLFromEnv())
}
