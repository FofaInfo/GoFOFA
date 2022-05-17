package gofofa

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_HostSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error
	var account accountInfo
	var res [][]string

	// 注册用户，没有F币
	account = validAccounts[0]
	cli, err = NewClient(ts.URL + "?email=" + account.Email + "&key=" + account.Key)
	assert.Nil(t, err)
	res, err = cli.HostSearch("port=80", 10, []string{"ip", "port"})
	assert.Contains(t, err.Error(), "insufficient privileges")
	// 注册用户，有F币
	account = validAccounts[4]
	cli, err = NewClient(ts.URL + "?email=" + account.Email + "&key=" + account.Key)
	assert.Nil(t, err)
	res, err = cli.HostSearch("port=80", 10, []string{"ip", "port"})
	assert.Contains(t, err.Error(), "DeductModeFCoin")

	// 参数错误
	account = validAccounts[1]
	cli, err = NewClient(ts.URL + "?email=" + account.Email + "&key=" + account.Key)
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	res, err = cli.HostSearch("", 10, []string{"ip", "port"})
	assert.Contains(t, err.Error(), "[-4] Params Error")
	assert.Equal(t, 0, len(res))

	// 数量超出限制
	res, err = cli.HostSearch("port=80", 10000, []string{"ip", "port"})
	assert.Equal(t, 100, len(res))
	account = validAccounts[2]
	cli, err = NewClient(ts.URL + "?email=" + account.Email + "&key=" + account.Key)
	res, err = cli.HostSearch("port=80", 10000, []string{"ip", "port"})
	assert.Equal(t, 10000, len(res))

	// 多字段
	account = validAccounts[1]
	cli, err = NewClient(ts.URL + "?email=" + account.Email + "&key=" + account.Key)
	res, err = cli.HostSearch("port=80", 10, []string{"ip", "port"})
	assert.Equal(t, 10, len(res))
	assert.Equal(t, "94.130.128.248", res[0][0])
	assert.Equal(t, "80", res[0][1])
	// 没有字段，跟ip，port一样
	res, err = cli.HostSearch("port=80", 10, nil)
	assert.Equal(t, "94.130.128.248", res[0][0])
	assert.Equal(t, "80", res[0][1])

	// 单字段
	res, err = cli.HostSearch("port=80", 10, []string{"host"})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res))

	// 0 数据
	res, err = cli.HostSearch("port=80", 0, nil)
	assert.Contains(t, err.Error(), "The Size value `0` must be between")
}

func TestClient_HostSize(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error
	var account accountInfo
	var count int

	account = validAccounts[1]
	cli, err = NewClient(ts.URL + "?email=" + account.Email + "&key=" + account.Key)
	assert.Nil(t, err)
	count, err = cli.HostSize("port=80")
	assert.Nil(t, err)
	assert.Equal(t, 12345678, count)
}
