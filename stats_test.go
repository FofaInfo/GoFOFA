package gofofa

import (
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestClient_Stats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error
	var account accountInfo
	var res []StatsObject

	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)

	// 错误
	res, err = cli.Stats("port=80", 0, []string{"title"})
	assert.Error(t, err)

	// 正确
	res, err = cli.Stats("port=80", 5, []string{"title"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 5, len(res[0].Items))
	assert.Equal(t, "title", res[0].Name)
	assert.Equal(t, 25983408, res[0].Items[0].Count)

	// 默认字段
	res, err = cli.Stats("port=80", 5, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, 5, len(res[0].Items))
	assert.Equal(t, "title", res[0].Name)
	assert.Equal(t, "301 Moved Permanently", res[0].Items[0].Name)
	assert.Equal(t, 25983454, res[0].Items[0].Count)
	assert.Equal(t, "country", res[1].Name)
	assert.Equal(t, "United States of America", res[1].Items[0].Name)
	assert.Equal(t, 154746752, res[1].Items[0].Count)

	// 请求失败
	cli = &Client{
		Server:     "http://fofa.info:66666",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	res, err = cli.Stats("port=80", 5, nil)
	assert.Error(t, err)
}

func TestClient_Stats_BuildUrl(t *testing.T) {
	fullURL := fmt.Sprintf("%s/api/%s/%s?", "https://fofa.info", "v1", "search/stats")
	params := map[string]string{
		"qbase64": base64.StdEncoding.EncodeToString([]byte("port=80")),
		"size":    strconv.Itoa(100),
		"fields":  "ip,port",
		"full":    "false", // 是否全部数据，非一年内
	}
	ps := url.Values{}
	ps.Set("email", "mayuze@baimaohui.net")
	ps.Set("key", "bf68530e5d5352ddb5f048de4d92e58b")
	for k, v := range params {
		ps.Set(k, v)
	}
	fmt.Println(fullURL + ps.Encode())
}
