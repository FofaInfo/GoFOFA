package gofofa

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/expr-lang/expr"
	"math"
	"strconv"
	"strings"
)

const (
	NoHostWithFixURL = "host field must included when fixUrl option set"
)

// HostResults /search/all api results
type HostResults struct {
	Mode    string      `json:"mode"`
	Error   bool        `json:"error"`
	Errmsg  string      `json:"errmsg"`
	Query   string      `json:"query"`
	Page    int         `json:"page"`
	Size    int         `json:"size"` // 总数
	Results interface{} `json:"results"`
	Next    string      `json:"next"`
}

// HostStatsData /host api results
type HostStatsData struct {
	Error       bool     `json:"error"`
	Errmsg      string   `json:"errmsg"`
	Host        string   `json:"host"`
	IP          string   `json:"ip"`
	ASN         int      `json:"asn"`
	ORG         string   `json:"org"`
	Country     string   `json:"country_name"`
	CountryCode string   `json:"country_code"`
	Protocols   []string `json:"protocol"`
	Ports       []int    `json:"port"`
	Categories  []string `json:"category"`
	Products    []string `json:"product"`
	UpdateTime  string   `json:"update_time"`
}

// SearchOptions options of search, for post processors
type SearchOptions struct {
	FixUrl     bool   // each host fix as url, like 1.1.1.1,80 will change to http://1.1.1.1, https://1.1.1.1:8443 will no change
	UrlPrefix  string // default is http://
	Full       bool   // search result for over a year
	UniqByIP   bool   // uniq by ip
	IsActive   int    // probe website is existed, add isActive field
	DeWildcard int    // number of wildcard domains retained
	Filter     string // filter data by rules
	DedupHost  bool   // prioritize subdomain data retention
}

// fixHostToUrl 替换host为url
func fixHostToUrl(res [][]string, fields []string, hostIndex int, urlPrefix string, protocolIndex int) [][]string {

	newRes := make([][]string, 0, len(res))
	for _, row := range res {
		newRow := make([]string, 0, len(fields))
		for j, r := range row {
			if j == hostIndex {
				if !strings.Contains(r, "://") {
					if urlPrefix != "" {
						r = urlPrefix + r
					} else if protocolIndex != -1 &&
						(row[protocolIndex] == "socks5" || row[protocolIndex] == "redis" ||
							row[protocolIndex] == "http" || row[protocolIndex] == "https" ||
							row[protocolIndex] == "mongodb" || row[protocolIndex] == "mysql") {
						r = row[protocolIndex] + "://" + r
					} else {
						r = "http://" + r
					}
				}
			}
			newRow = append(newRow, r)
		}
		newRes = append(newRes, newRow)
	}
	return newRes
}

func getParamIndexThenAdd(fields []string, field string) (int, []string) {
	paramIndex := -1
	for index, f := range fields {
		if f == field {
			paramIndex = index
			break
		}
	}
	if paramIndex == -1 {
		fields = append(fields, field)
		paramIndex = len(fields) - 1
	}
	return paramIndex, fields
}

func extractVariables(filter string) ([]string, error) {
	f, err := govaluate.NewEvaluableExpression(filter)
	if err != nil {
		return nil, err
	}

	variables := f.Vars()
	return variables, nil
}

// fixUrlCheck 检查参数，构建新的field和记录相关字段的偏移
// 返回hostIndex, protocolIndex, fields, rawFieldSize, err
func (c *Client) fixUrlCheck(fields []string, options ...SearchOptions) (int, int, []string, int, error) {
	noSetFields := false
	if len(fields) == 0 {
		noSetFields = true
		fields = []string{"host", "ip", "port"}
	}
	rawFieldSize := len(fields)

	// 确保urlfix开启后带上了protocol字段
	protocolIndex := -1
	hostIndex := -1
	if len(options) > 0 && options[0].FixUrl {
		if noSetFields {
			fields = []string{"host", "ip", "port", "protocol"}
			rawFieldSize = len(fields)
			hostIndex = 0
			protocolIndex = 3
		} else {
			// 检查host字段存在
			for index, f := range fields {
				switch f {
				case "host":
					hostIndex = index
					break
				}
			}
			if hostIndex == -1 {
				err := errors.New(NoHostWithFixURL)
				return hostIndex, protocolIndex, fields, rawFieldSize, err
			}
			for index, f := range fields {
				switch f {
				case "protocol":
					protocolIndex = index
					break
				}
			}
			if protocolIndex == -1 {
				fields = append(fields, "protocol")
				protocolIndex = len(fields) - 1
			}
		}
	}
	return hostIndex, protocolIndex, fields, rawFieldSize, nil
}

func (c *Client) postProcess(res [][]string, fields []string,
	hostIndex int, protocolIndex int, rawFieldSize int, options ...SearchOptions) [][]string {
	if len(options) > 0 && options[0].FixUrl {
		res = fixHostToUrl(res, fields, hostIndex, options[0].UrlPrefix, protocolIndex)
	}

	// 返回用户指定的字段
	if rawFieldSize != len(fields) {
		var newRes [][]string
		for _, r := range res {
			newRes = append(newRes, r[0:rawFieldSize])
		}
		return newRes
	}
	return res
}

// HostSearch search fofa host data
// query fofa query string
// size data size: -1 means all，0 means just data total info, >0 means actual size
// fields of fofa host search
// options for search
func (c *Client) HostSearch(query string, size int, fields []string, options ...SearchOptions) (res [][]string, err error) {
	var (
		full        bool
		uniqByIP    bool
		isActive    int
		dedupCname  int
		isSubDomain bool
		filter      string
	)
	if len(options) > 0 {
		full = options[0].Full
		uniqByIP = options[0].UniqByIP
		isActive = options[0].IsActive
		dedupCname = options[0].DeWildcard
		filter = options[0].Filter
		isSubDomain = options[0].DedupHost
	}

	freeSize := c.freeSize()
	// check level
	if freeSize == 0 {
		// 不是会员
		if c.Account.FCoin < 1 {
			return nil, errors.New("insufficient privileges") // 等级不够，fcoin也不够
		}
		if c.DeductMode != DeductModeFCoin {
			return nil, errors.New("insufficient privileges, try to set mode to 1(DeductModeFCoin)") // 等级不够，fcoin也不够
		}
	} else if freeSize == -1 {
		// unknown vip level, skip mode check
	} else if size > c.freeSize() {
		// 是会员，但是取的数量比免费的大
		switch c.DeductMode {
		case DeductModeFree:
			// 防止 freesize = -1，取 size 和 freesize 的最大值
			if freeSize <= 0 {
				size = int(math.Max(float64(freeSize), float64(size)))
			} else {
				size = freeSize
			}
			c.logger.Warnf("size is larger than your account free limit, "+
				"just fetch %d instead, if you want deduct fcoin automatically, set mode to 1(DeductModeFCoin) manually", size)
		}
	}

	page := 1
	perPage := int(math.Min(float64(size), 1000)) // 最多一次取1000

	// 一次取所有数据，perPage 默认给 1000
	if size == -1 {
		perPage = 1000
	}

	hostIndex, protocolIndex, fields, rawFieldSize, err := c.fixUrlCheck(fields, options...)
	if err != nil {
		return nil, err
	}

	uniqIPMap := make(map[string]bool)
	// 确认fields包含ip
	var ipIndex = -1
	if uniqByIP {
		ipIndex, fields = getParamIndexThenAdd(fields, "ip")
	}

	var activeSlice []string
	// 确认fields包含link
	var linkIndex, codeIndex = -1, -1
	if isActive > 0 {
		linkIndex, fields = getParamIndexThenAdd(fields, "link")
		codeIndex, fields = getParamIndexThenAdd(fields, "status_code")
	}

	dedupCnameMap := make(map[string]int)
	// 确认fields包含ip、port、domain、title、fid
	var portIndex, domainIndex, titleIndex, fidIndex int = -1, -1, -1, -1
	if dedupCname > 0 {
		ipIndex, fields = getParamIndexThenAdd(fields, "ip")
		portIndex, fields = getParamIndexThenAdd(fields, "port")
		domainIndex, fields = getParamIndexThenAdd(fields, "domain")
		titleIndex, fields = getParamIndexThenAdd(fields, "title")
		fidIndex, fields = getParamIndexThenAdd(fields, "fid")
	}

	// 过滤器配置
	filterIndexs := make(map[string]int)
	if len(filter) > 0 {
		var variables []string
		variables, err = extractVariables(filter)
		if err != nil {
			return nil, err
		}

		var filterIndex = -1
		for _, filterField := range variables {
			filterIndex, fields = getParamIndexThenAdd(fields, filterField)
			filterIndexs[filterField] = filterIndex
		}
	}

	isSubDomainMap := make(map[string][]string)
	// 确认fields包含type
	typeIndex := -1
	if isSubDomain {
		typeIndex, fields = getParamIndexThenAdd(fields, "type")
		linkIndex, fields = getParamIndexThenAdd(fields, "link")
	}

	// 分页取数据
	for {
		if ctx := c.GetContext(); ctx != nil {
			// 确认是否需要退出
			select {
			case <-c.GetContext().Done():
				err = context.Canceled
				return
			default:
			}
		}

		var hr HostResults
		err = c.Fetch("search/all",
			map[string]string{
				"qbase64": base64.StdEncoding.EncodeToString([]byte(query)),
				"size":    strconv.Itoa(perPage),
				"page":    strconv.Itoa(page),
				"fields":  strings.Join(fields, ","),
				"full":    strconv.FormatBool(full), // 是否全部数据，非一年内
			},
			&hr)
		if err != nil {
			return
		}

		// 报错，退出
		if len(hr.Errmsg) > 0 {
			err = errors.New(hr.Errmsg)
			break
		}

		var results [][]string
		if v, ok := hr.Results.([]interface{}); ok {
			// 无数据
			if len(v) == 0 {
				break
			}
			for _, result := range v {
				if vStrSlice, ok := result.([]interface{}); ok {
					var newSlice []string
					for _, vStr := range vStrSlice {
						newSlice = append(newSlice, vStr.(string))
					}
					if uniqByIP {
						if _, ok := uniqIPMap[newSlice[ipIndex]]; ok {
							continue
						}
						uniqIPMap[newSlice[ipIndex]] = true
					}
					if dedupCname > 0 {
						key := fmt.Sprintf("%s:%s:%s:%s:%s", newSlice[ipIndex], newSlice[portIndex],
							newSlice[domainIndex], newSlice[titleIndex], newSlice[fidIndex])
						if _, ok := dedupCnameMap[key]; ok && dedupCnameMap[key] > 3 {
							continue
						}
						dedupCnameMap[key]++
					}
					if len(filter) > 0 {
						env := make(map[string]interface{})
						for field, index := range filterIndexs {
							env[field] = newSlice[index]
						}

						program, err := expr.Compile(filter, expr.Env(env))
						if err != nil {
							return nil, err
						}

						match, err := expr.Run(program, env)
						if err != nil {
							return nil, err
						}

						if !match.(bool) {
							continue
						}
					}
					if isActive > 0 {
						resp := DoHttpCheck(newSlice[linkIndex], isActive)
						activeSlice = append(activeSlice, fmt.Sprintf("%t", resp.IsActive))
						newSlice[codeIndex] = resp.StatusCode
					}
					results = append(results, newSlice)
				} else if vStr, ok := result.(string); ok {
					// 确定第一个就是ip
					newSlice := []string{vStr}
					if uniqByIP && ipIndex == 0 {
						if _, ok := uniqIPMap[vStr]; ok {
							continue
						}
						uniqIPMap[vStr] = true
					}
					if isActive > 0 && linkIndex == 0 {
						resp := DoHttpCheck(vStr, isActive)
						activeSlice = append(activeSlice, fmt.Sprintf("%t", resp.IsActive))
					}
					results = append(results, newSlice)
				}
			}
		} else {
			break
		}

		if c.onResults != nil {
			c.onResults(results)
		}

		res = append(res, results...)

		// 数据填满了，完成
		if size != -1 && size <= len(res) {
			break
		}

		// 数据已经没有了
		if len(hr.Results.([]interface{})) < perPage {
			break
		}

		page++ // 翻页
	}

	if isSubDomain {
		var result [][]string
		for _, row := range res {
			exist, found := isSubDomainMap[row[linkIndex]]
			if found {
				if row[linkIndex] == "" {
					result = append(result, row)
					continue
				}
				if !(exist[typeIndex] == "service" && row[typeIndex] == "subdomain") {
					continue
				}
			}
			isSubDomainMap[row[linkIndex]] = row
		}

		for _, v := range isSubDomainMap {
			result = append(result, v)
		}

		res = result
	}

	// 后处理
	res = c.postProcess(res, fields, hostIndex, protocolIndex, rawFieldSize, options...)
	if isActive > 0 {
		for index := range res {
			res[index] = append(res[index], activeSlice[index])
		}
	}

	return
}

// HostSize fetch query matched host count
func (c *Client) HostSize(query string) (count int, err error) {
	var hr HostResults
	err = c.Fetch("search/all",
		map[string]string{
			"qbase64": base64.StdEncoding.EncodeToString([]byte(query)),
			"size":    "1",
			"page":    "1",
			"full":    "false", // 是否全部数据，非一年内
		},
		&hr)
	if err != nil {
		return
	}
	count = hr.Size
	return
}

// HostStats fetch query matched host count
func (c *Client) HostStats(host string) (data HostStatsData, err error) {
	err = c.Fetch("host/"+host, nil, &data)
	if err != nil {
		return
	}
	if data.Error {
		err = errors.New(data.Errmsg)
	}
	return
}

// DumpSearch search fofa host data
// query fofa query string
// size data size: -1 means all，0 means just data total info, >0 means actual size
// fields of fofa host search
// options for search
func (c *Client) DumpSearch(query string, allSize int, batchSize int, fields []string, onResults func([][]string, int) error, options ...SearchOptions) (err error) {
	var full bool
	if len(options) > 0 {
		full = options[0].Full
	}

	next := ""
	perPage := batchSize
	if perPage < 1 || perPage > 100000 {
		return errors.New("batchSize must between 1 and 100000")
	}

	// 确保urlfix开启后带上了protocol字段
	hostIndex, protocolIndex, fields, rawFieldSize, err := c.fixUrlCheck(fields, options...)
	if err != nil {
		return err
	}

	// 分页取数据
	fetchedSize := 0
	for {
		if ctx := c.GetContext(); ctx != nil {
			// 确认是否需要退出
			select {
			case <-c.GetContext().Done():
				err = context.Canceled
				return
			default:
			}
		}

		var hr HostResults
		err = c.Fetch("search/next",
			map[string]string{
				"qbase64": base64.StdEncoding.EncodeToString([]byte(query)),
				"size":    strconv.Itoa(perPage),
				"fields":  strings.Join(fields, ","),
				"full":    strconv.FormatBool(full), // 是否全部数据，非一年内
				"next":    next,                     // 偏移
			},
			&hr)
		if err != nil {
			return
		}

		// 报错，退出
		if len(hr.Errmsg) > 0 {
			err = errors.New(hr.Errmsg)
			break
		}

		var results [][]string
		if v, ok := hr.Results.([]interface{}); ok {
			// 无数据
			if len(v) == 0 {
				break
			}
			for _, result := range v {
				if vStrSlice, ok := result.([]interface{}); ok {
					var newSlice []string
					for _, vStr := range vStrSlice {
						newSlice = append(newSlice, vStr.(string))
					}
					results = append(results, newSlice)
				} else if vStr, ok := result.(string); ok {
					results = append(results, []string{vStr})
				}
			}
		} else {
			break
		}

		// 后处理
		results = c.postProcess(results, fields, hostIndex, protocolIndex, rawFieldSize, options...)

		if c.onResults != nil {
			c.onResults(results)
		}
		if err := onResults(results, hr.Size); err != nil {
			return err
		}

		fetchedSize += len(results)

		// 数据填满了，完成
		if allSize > 0 && allSize <= fetchedSize {
			break
		}

		// 数据已经没有了
		if len(results) < perPage {
			break
		}

		// 结束
		if hr.Next == "" {
			break
		}

		next = hr.Next // 偏移
	}

	return
}
