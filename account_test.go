package gofofa

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestParseDeductMode(t *testing.T) {
	assertPanic(t, func() {
		ParseDeductMode("abc")
	})

	assert.Equal(t, DeductModeFree, ParseDeductMode("DeductModeFree"))
	assert.Equal(t, DeductModeFree, ParseDeductMode("0"))
	assert.Equal(t, DeductModeFCoin, ParseDeductMode("DeductModeFCoin"))
	assert.Equal(t, DeductModeFCoin, ParseDeductMode("1"))
}

func TestAccountInfo_String(t *testing.T) {
	ai := AccountInfo{
		Error:    false,
		IsVIP:    true,
		VIPLevel: 3,
		FCoin:    0,
	}
	assert.EqualValues(t, AccountInfo{
		Error:          false,
		FCoin:          0,
		FofaPoint:      0,
		IsVIP:          true,
		VIPLevel:       3,
		RemainApiQuery: 0,
		RemainApiData:  0,
	}, ai)
	assert.Equal(t, `{
  "error": false,
  "fcoin": 0,
  "fofa_point": 0,
  "isvip": true,
  "vip_level": 3,
  "remain_api_query": 0,
  "remain_api_data": 0
}`, ai.String())
}

func TestClient_AccountInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error

	// 请求失败
	errTs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer ts.Close()
	_, err = NewClient(WithURL(errTs.URL + "?email=a@a.com&key=wrong"))
	assert.EqualError(t, err, "unexpected end of JSON input")

	// 账号无效
	_, err = NewClient(WithURL(ts.URL + "?email=a@a.com&key=wrong"))
	assert.Contains(t, err.Error(), "[-700] Account Invalid")

	// 注册用户
	var account accountInfo
	account = validAccounts[0]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.False(t, cli.Account.IsVIP)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, 0, cli.freeSize())

	// 普通会员
	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelNormal, cli.Account.VIPLevel)
	assert.Equal(t, 10, cli.Account.FCoin)
	assert.Equal(t, 100, cli.freeSize())

	// 高级会员
	account = validAccounts[2]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelAdvanced, cli.Account.VIPLevel)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, 10000, cli.freeSize())

	// 企业会员
	account = validAccounts[3]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelEnterprise, cli.Account.VIPLevel)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, 100000, cli.freeSize())

	// 订阅个人
	account = validAccounts[5]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelSubPersonal, cli.Account.VIPLevel)
	assert.Equal(t, 10, cli.Account.FCoin)
	assert.Equal(t, 100, cli.freeSize())

	// 订阅商业版
	account = validAccounts[7]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelSubBuss, cli.Account.VIPLevel)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, 100000, cli.freeSize())
	// 构造异常
	cli.Server = errTs.URL
	assert.Equal(t, 100000, cli.freeSize())

	// 红队？
	account = validAccounts[8]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelRed, cli.Account.VIPLevel)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, 10000, cli.freeSize())

	// 学生
	account = validAccounts[9]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelStudent, cli.Account.VIPLevel)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, 10000, cli.freeSize())

	// 不可能的等级
	account = validAccounts[10]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	assert.Equal(t, VipLevelNever, cli.Account.VIPLevel)
	assert.Equal(t, 0, cli.Account.FCoin)
	assert.Equal(t, -1, cli.freeSize())
}

func TestClient_AccountInfoReturnsErrorFlagWithoutMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{"error": true})
	}))
	defer server.Close()

	client := &Client{
		Server:     server.URL,
		APIVersion: "v1",
		httpClient: server.Client(),
	}
	_, err := client.AccountInfo()
	assert.EqualError(t, err, "fofa account info failed")
}
