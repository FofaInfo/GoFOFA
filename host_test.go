package gofofa

import (
	"context"
	"github.com/sirupsen/logrus"
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
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	res, err = cli.HostSearch("port=80", 10, []string{"ip", "port"})
	assert.Contains(t, err.Error(), "insufficient privileges")
	// 注册用户，有F币
	account = validAccounts[4]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	res, err = cli.HostSearch("port=80", 10, []string{"ip", "port"})
	assert.Contains(t, err.Error(), "DeductModeFCoin")

	// 参数错误
	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	assert.True(t, cli.Account.IsVIP)
	res, err = cli.HostSearch("", 10, []string{"ip", "port"})
	assert.Contains(t, err.Error(), "[-4] Params Error")
	assert.Equal(t, 0, len(res))

	// 数量超出限制
	res, err = cli.HostSearch("port=80", 10000, []string{"ip", "port"})
	assert.Equal(t, 100, len(res))
	account = validAccounts[2]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	res, err = cli.HostSearch("port=80", 10000, []string{"ip", "port"})
	assert.Equal(t, 10000, len(res))

	// 多字段
	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	res, err = cli.HostSearch("port=80", 10, []string{"ip", "port"})
	assert.Equal(t, 10, len(res))
	assert.Equal(t, "94.130.128.248", res[0][0])
	assert.Equal(t, "80", res[0][1])
	// 没有字段，跟host, ip，port一样
	res, err = cli.HostSearch("port=80", 10, nil)
	assert.Equal(t, "1.1.1.1:81", res[0][0])
	assert.Equal(t, "1.1.1.1", res[0][1])
	assert.Equal(t, "81", res[0][2])

	// 单字段
	res, err = cli.HostSearch("port=80", 10, []string{"host"})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res))

	// 请求0数据
	res, err = cli.HostSearch("port=80", 0, nil)
	assert.Contains(t, err.Error(), "The Size value `0` must be between")

	// 返回0条数据
	res, err = cli.HostSearch("port=100000", 10, nil)
	assert.Nil(t, err)
	assert.Nil(t, res)

	// 返回非正常格式数据
	res, err = cli.HostSearch("port=100001", 10, nil)
	assert.Nil(t, err)

	// 数据不够
	res, err = cli.HostSearch("port=50000", 10000, nil)
	assert.Nil(t, err)
	assert.Equal(t, 9, len(res))

	// 错误语句
	res, err = cli.HostSearch("aaa=bbb", 10, nil)
	assert.Contains(t, err.Error(), "[820000] FOFA Query Syntax Incorrect")

	// search full result
	res, err = cli.HostSearch("port=5354", 100, []string{"ip", "port"}, SearchOptions{
		Full: false,
	})
	assert.Nil(t, err)
	res2, err := cli.HostSearch("port=5354", 100, []string{"ip", "port"}, SearchOptions{
		Full: true,
	})
	assert.Nil(t, err)
	assert.Greater(t, len(res2), len(res))

	// 没有权限
	res, err = cli.HostSearch("port=1231", 10, []string{"fid"})
	assert.Contains(t, err.Error(), "没有权限搜索fid字段")

	// 带有fixurl
	res, err = cli.HostSearch("port=80", 10, []string{"host"}, SearchOptions{
		FixUrl:    true,
		UrlPrefix: "",
	})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res))
	assert.Contains(t, res[0][0], "https://")
	res, err = cli.HostSearch("port=80", 10, []string{"host"}, SearchOptions{
		FixUrl:    true,
		UrlPrefix: "redis://",
	})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res))
	assert.Contains(t, res[0][0], "redis://")

	// 请求失败
	cli = &Client{
		Server:     "http://fofa.info:66666",
		httpClient: &http.Client{},
		Account: AccountInfo{
			FCoin:    0,
			IsVIP:    true,
			VIPLevel: 1,
		},
		logger: logrus.New(),
	}
	res, err = cli.HostSearch("port=80", 10, []string{"host"})
	assert.Error(t, err)

	// search all data
	account = validAccounts[3]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))

	res, err = cli.HostSearch("port=80", -1, []string{"host"}, SearchOptions{
		FixUrl: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res))

	// 过滤UniqByIP
	res, err = cli.HostSearch("port=80", 500, []string{"ip", "port"}, SearchOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 1000, len(res))

	res, err = cli.HostSearch("port=80", 500, []string{"ip", "port"}, SearchOptions{
		UniqByIP: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, 986, len(res))

	// 域名泛解析保留
	res, err = cli.HostSearch("domain=huashunxinan.net", 500, []string{"link"}, SearchOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 100, len(res))
	res, err = cli.HostSearch("domain=huashunxinan.net", 500, []string{"link"}, SearchOptions{
		DeWildcard: 3,
	})
	assert.Nil(t, err)
	assert.Equal(t, 11, len(res))

	// subdomain去重
	res, err = cli.HostSearch("domain=huashunxinan.net", 500, []string{"link"}, SearchOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 100, len(res))
	res, err = cli.HostSearch("domain=huashunxinan.net", 500, []string{"link"}, SearchOptions{
		DedupHost: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, 99, len(res))

	// checkActive探活
	res, err = cli.HostSearch("ip=127.0.0.1", 1, []string{"link"}, SearchOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res[0]))
	res, err = cli.HostSearch("ip=127.0.0.1", 1, []string{"link"}, SearchOptions{
		CheckActive: 3,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res[0])) // 多一个isActive字段

	// filter过滤器
	res, err = cli.HostSearch("domain=huashunxinan.net", 500, []string{"ip", "host", "status_code"}, SearchOptions{})
	assert.Nil(t, err)
	assert.Equal(t, 100, len(res))
	res, err = cli.HostSearch("domain=huashunxinan.net", 500, []string{"ip", "host", "status_code"}, SearchOptions{
		Filter: "status_code=='200'",
	})
	assert.Nil(t, err)
	assert.Equal(t, 57, len(res))
}

func TestClient_HostSearch_UniqIP(t *testing.T) {
	fofaQueryTest := `title=test`
	ts := httptest.NewServer(http.HandlerFunc(bindSearchAllQueryHandle(fofaQueryTest, "ip,port",
		`{"error":false,"size":3,"page":1,"mode":"extended","query":"title=\"test\"","results":[["1.1.1.1","80"],["1.1.1.1","81"],["2.2.2.2","81"]]}`,
	)))
	defer ts.Close()
	var cli *Client
	var err error
	var account accountInfo
	var res [][]string

	account = validAccounts[3]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))

	// fixUrl support socks5/redis
	res, err = cli.HostSearch(fofaQueryTest, 10, []string{"ip", "port"}, SearchOptions{
		UniqByIP: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, "1.1.1.1", res[0][0])
	assert.Equal(t, "2.2.2.2", res[1][0])
}

func TestClient_HostSearch_FixUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()
	var cli *Client
	var err error
	var account accountInfo
	var res [][]string

	account = validAccounts[3]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))

	// fixUrl support socks5/redis
	res, err = cli.HostSearch("port=1080", 10, []string{"ip", "port"}, SearchOptions{
		FixUrl: true,
	})
	assert.EqualError(t, err, NoHostWithFixURL)

	// fixUrl support socks5/redis
	res, err = cli.HostSearch("port=1080", 10, []string{"host", "ip", "port"}, SearchOptions{
		FixUrl: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, "socks5://1.1.1.1:1080", res[0][0])
	assert.Equal(t, "redis://2.2.2.2:1080", res[1][0])
	assert.Equal(t, "https://3.3.3.3:1080", res[2][0])

	// fixUrl support socks5/redis
	res, err = cli.HostSearch("port=1080", 10, []string{"host", "ip", "port"}, SearchOptions{
		FixUrl:    true,
		UrlPrefix: "",
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, "socks5://1.1.1.1:1080", res[0][0])
	assert.Equal(t, "redis://2.2.2.2:1080", res[1][0])
	assert.Equal(t, "https://3.3.3.3:1080", res[2][0])

	// no fields
	res, err = cli.HostSearch("port=1080", 10, nil, SearchOptions{
		FixUrl:    true,
		UrlPrefix: "",
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, "socks5://1.1.1.1:1080", res[0][0])
	assert.Equal(t, "redis://2.2.2.2:1080", res[1][0])
	assert.Equal(t, "https://3.3.3.3:1080", res[2][0])

	// has protocol
	res, err = cli.HostSearch("port=1080", 10, []string{"host", "ip", "port", "protocol"}, SearchOptions{
		FixUrl:    true,
		UrlPrefix: "",
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res))
	assert.Equal(t, "socks5://1.1.1.1:1080", res[0][0])
	assert.Equal(t, "redis://2.2.2.2:1080", res[1][0])
	assert.Equal(t, "https://3.3.3.3:1080", res[2][0])

	// no url prefix and no protocol
	hosts := fixHostToUrl([][]string{{"1.1.1.1:80", "1.1.1.1", "80"}}, []string{"host", "ip", "port"}, 0, "", -1)
	assert.Equal(t, "http://1.1.1.1:80", hosts[0][0])

	// todo：确保返回的字段数跟用户要求的一致

}

func TestClient_HostSize(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error
	var account accountInfo
	var count int

	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	count, err = cli.HostSize("port=80")
	assert.Nil(t, err)
	assert.Equal(t, 12345678, count)

	// 请求失败
	cli = &Client{
		Server:     "http://fofa.info:66666",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	count, err = cli.HostSize("port=80")
	assert.Error(t, err)
}

func TestClient_HostStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error
	var account accountInfo
	var hostStat HostStatsData

	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	assert.Nil(t, err)
	hostStat, err = cli.HostStats("1.1.1.1")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(hostStat.Ports))
	hostStat, err = cli.HostStats("fofa.info")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(hostStat.Ports))

	// 请求失败
	cli = &Client{
		Server:     "http://fofa.info:66666",
		httpClient: &http.Client{},
		logger:     logrus.New(),
	}
	hostStat, err = cli.HostStats("1.1.1.1")
	assert.Error(t, err)
}

func TestClient_SetContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	onDataCh := make(chan struct{}, 1)
	onResults := func([][]string) {
		onDataCh <- struct{}{}
	}
	account := validAccounts[3]
	cli, err := NewClient(WithURL(ts.URL+"?email="+account.Email+"&key="+account.Key), WithOnResults(onResults))
	assert.Nil(t, err)
	cli.DeductMode = DeductModeFCoin

	ctx, cancel := context.WithCancel(context.Background())
	cli.SetContext(ctx)
	stopCh := make(chan struct{}, 1)
	go func() {
		defer func() {
			stopCh <- struct{}{}
		}()
		res, err := cli.HostSearch("port=80", 100000000, []string{"ip", "port"})
		assert.Equal(t, context.Canceled, err)
		assert.True(t, len(res) > 0)
	}()
	<-onDataCh
	cancel()
	<-stopCh
}

func TestClient_DumpSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(queryHander))
	defer ts.Close()

	var cli *Client
	var err error
	var account accountInfo
	var res [][]string

	// 多字段
	account = validAccounts[1]
	cli, err = NewClient(WithURL(ts.URL + "?email=" + account.Email + "&key=" + account.Key))
	err = cli.DumpSearch("port=80", 10000, 10, []string{"ip", "port"}, func(i [][]string, i2 int) error {
		res = append(res, i...)
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 100, len(res))
	assert.Equal(t, "1.1.1.1", res[0][0])
	assert.Equal(t, "81", res[0][1])

	res = nil
	err = cli.DumpSearch("port=80", 10000, 10, nil, func(i [][]string, i2 int) error {
		res = append(res, i...)
		return nil
	}, SearchOptions{FixUrl: true})
	assert.Nil(t, err)
	assert.Equal(t, 100, len(res))
	assert.Equal(t, "http://1.1.1.1", res[0][0])
	assert.Equal(t, "1.1.1.1", res[0][1])
	assert.Equal(t, "81", res[0][2])

	// 数据范围报错
	res = nil
	err = cli.DumpSearch("port=80", 10000, 10000000, nil, func(i [][]string, i2 int) error {
		res = append(res, i...)
		return nil
	}, SearchOptions{FixUrl: true})
	assert.NotNil(t, err)

	// 取消
	ctx, cancel := context.WithCancel(context.Background())
	cli.SetContext(ctx)
	res = nil
	err = cli.DumpSearch("port=80", 10000, 10, nil, func(i [][]string, i2 int) error {
		cancel()
		return nil
	}, SearchOptions{FixUrl: true})
	assert.NotNil(t, err)
}
