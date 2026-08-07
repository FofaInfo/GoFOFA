package gofofa

import "encoding/json"

// DeductMode should deduct fcoin automatically or just use free limit
type DeductMode int

const (
	// DeductModeFree only use free limit size
	DeductModeFree DeductMode = 0
	// DeductModeFCoin deduct fcoin automatically if account has fcoin
	DeductModeFCoin DeductMode = 1
)

// ParseDeductMode parse string to DeductMode
func ParseDeductMode(v string) DeductMode {
	switch v {
	case "0", "DeductModeFree":
		return DeductModeFree
	case "1", "DeductModeFCoin":
		return DeductModeFCoin
	default:
		panic("unknown deduct mode")
	}
}

type VipLevel int

const (
	VipLevelNone        VipLevel = 0 // 注册用户
	VipLevelNormal      VipLevel = 1 // 普通会员
	VipLevelAdvanced    VipLevel = 2 // 高级会员
	VipLevelEnterprise  VipLevel = 3 // 企业版
	VipLevelEnterprise2 VipLevel = 5 // 企业版
)

const (
	VipLevelSubPersonal VipLevel = 11 // 订阅个人
	VipLevelSubPro      VipLevel = 12 // 订阅专业
	VipLevelSubBuss     VipLevel = 13 // 订阅商业版
)

const (
	VipLevelRed     VipLevel = 20  // 红队版
	VipLevelStudent VipLevel = 22  // 教育账户
	VipLevelNever   VipLevel = 100 // 不可能的等级
)

// AccountInfo fofa account info
type AccountInfo struct {
	Error          bool     `json:"error"`            // error or not
	ErrMsg         string   `json:"errmsg,omitempty"` // error string message
	FCoin          int      `json:"fcoin"`            // fcoin count
	FofaPoint      int64    `json:"fofa_point"`       // fofa point
	IsVIP          bool     `json:"isvip"`            // is vip
	VIPLevel       VipLevel `json:"vip_level"`        // vip level
	RemainApiQuery int      `json:"remain_api_query"` // available query
	RemainApiData  int      `json:"remain_api_data"`  // available data amount
}

func (ai AccountInfo) String() string {
	d, _ := json.MarshalIndent(ai, "", "  ")
	return string(d)
}

// AccountInfo fetch account info from fofa
func (c *Client) AccountInfo() (ac AccountInfo, err error) {
	err = c.Fetch("info/my", nil, &ac)
	if err != nil {
		return
	}
	err = apiResponseError(ac.Error, ac.ErrMsg, "fofa account info failed")
	return
}

// freeSize 获取可以免费使用的数据量
func (c *Client) freeSize() int {
	if !c.Account.IsVIP {
		// 不是会员有
		return 0
	}

	switch c.Account.VIPLevel {
	//case 0: // 上面已经退出了
	//	return 0
	case VipLevelNormal:
		return 100
	case VipLevelAdvanced:
		return 10000
	case VipLevelEnterprise, VipLevelEnterprise2:
		return 100000
	case VipLevelRed:
		return 10000
	case VipLevelStudent:
		return 10000
	// 订阅用户：通过 api 查询余额
	case VipLevelSubPersonal:
		fallthrough
	case VipLevelSubPro:
		fallthrough
	case VipLevelSubBuss:
		info, err := c.AccountInfo()
		if err != nil {
			info = c.Account
		}
		if info.RemainApiQuery > 0 {
			return info.RemainApiData
		}
	}
	// other level, ignore free limit check
	return -1
}
