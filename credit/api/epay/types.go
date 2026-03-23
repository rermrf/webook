package epay

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrInvalidSign      = errors.New("签名错误")
	ErrInvalidPid       = errors.New("商户ID错误")
	ErrInvalidOrder     = errors.New("订单不存在")
	ErrOrderPaid        = errors.New("订单已支付")
	ErrInsufficientBal  = errors.New("积分余额不足")
	ErrInvalidNotifyURL = errors.New("通知地址无效")
)

// 支付类型（积分支付只有一种类型）
const (
	PayTypeCredit = "credit" // 积分支付
)

// 交易状态
const (
	TradeStatusWait    = "WAIT_PAY"      // 待支付
	TradeStatusSuccess = "TRADE_SUCCESS" // 支付成功
	TradeStatusClosed  = "TRADE_CLOSED"  // 已关闭
)

// SubmitRequest 发起支付请求
type SubmitRequest struct {
	Pid        string `form:"pid" binding:"required"`          // 商户ID（对应app_id）
	Type       string `form:"type"`                            // 支付类型（默认credit）
	OutTradeNo string `form:"out_trade_no" binding:"required"` // 商户订单号
	NotifyURL  string `form:"notify_url" binding:"required"`   // 异步通知地址
	ReturnURL  string `form:"return_url"`                      // 同步跳转地址
	Name       string `form:"name" binding:"required"`         // 商品名称
	Money      string `form:"money" binding:"required"`        // 金额（积分数量）
	Param      string `form:"param"`                           // 自定义参数（原样返回）
	Sign       string `form:"sign" binding:"required"`         // 签名
	SignType   string `form:"sign_type"`                       // 签名类型（默认MD5）

	// 扩展字段
	Uid int64 `form:"uid" binding:"required"` // 支付用户ID
}

// SubmitResponse 发起支付响应（重定向到支付页面或返回错误）
type SubmitResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	TradeNo string `json:"trade_no,omitempty"` // 平台订单号
}

// QueryRequest 查询订单请求
type QueryRequest struct {
	Pid        string `form:"pid" binding:"required"`  // 商户ID
	OutTradeNo string `form:"out_trade_no"`            // 商户订单号（二选一）
	TradeNo    string `form:"trade_no"`                // 平台订单号（二选一）
	Sign       string `form:"sign" binding:"required"` // 签名
	SignType   string `form:"sign_type"`               // 签名类型
}

// QueryResponse 查询订单响应
type QueryResponse struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	TradeNo     string `json:"trade_no,omitempty"`      // 平台订单号
	OutTradeNo  string `json:"out_trade_no,omitempty"`  // 商户订单号
	Type        string `json:"type,omitempty"`          // 支付类型
	Pid         string `json:"pid,omitempty"`           // 商户ID
	Name        string `json:"name,omitempty"`          // 商品名称
	Money       string `json:"money,omitempty"`         // 金额
	TradeStatus string `json:"trade_status,omitempty"`  // 交易状态
	Param       string `json:"param,omitempty"`         // 自定义参数
	Addtime     string `json:"addtime,omitempty"`       // 创建时间
	Endtime     string `json:"endtime,omitempty"`       // 完成时间
}

// NotifyParams 异步通知参数
type NotifyParams struct {
	Pid         string `json:"pid"`          // 商户ID
	TradeNo     string `json:"trade_no"`     // 平台订单号
	OutTradeNo  string `json:"out_trade_no"` // 商户订单号
	Type        string `json:"type"`         // 支付类型
	Name        string `json:"name"`         // 商品名称
	Money       string `json:"money"`        // 金额
	TradeStatus string `json:"trade_status"` // 交易状态
	Param       string `json:"param"`        // 自定义参数
	Sign        string `json:"sign"`         // 签名
	SignType    string `json:"sign_type"`    // 签名类型
}

// Sign 生成签名
// 易支付签名规则：参数按key排序，拼接成 key1=value1&key2=value2，末尾加上密钥，然后MD5
func Sign(params map[string]string, key string) string {
	// 过滤空值和签名字段
	filtered := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filtered[k] = v
		}
	}

	// 按key排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(filtered[k])
	}
	buf.WriteString(key)

	// MD5
	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}

// VerifySign 验证签名
func VerifySign(params map[string]string, key, sign string) bool {
	expected := Sign(params, key)
	return strings.EqualFold(expected, sign)
}

// ToMap 将SubmitRequest转为map用于签名
func (r *SubmitRequest) ToMap() map[string]string {
	m := map[string]string{
		"pid":          r.Pid,
		"type":         r.Type,
		"out_trade_no": r.OutTradeNo,
		"notify_url":   r.NotifyURL,
		"return_url":   r.ReturnURL,
		"name":         r.Name,
		"money":        r.Money,
		"param":        r.Param,
	}
	// 添加扩展字段
	if r.Uid > 0 {
		m["uid"] = strconv.FormatInt(r.Uid, 10)
	}
	return m
}

// ToMap 将QueryRequest转为map用于签名
func (r *QueryRequest) ToMap() map[string]string {
	return map[string]string{
		"pid":          r.Pid,
		"out_trade_no": r.OutTradeNo,
		"trade_no":     r.TradeNo,
	}
}

// ToMap 将NotifyParams转为map用于签名
func (n *NotifyParams) ToMap() map[string]string {
	return map[string]string{
		"pid":          n.Pid,
		"trade_no":     n.TradeNo,
		"out_trade_no": n.OutTradeNo,
		"type":         n.Type,
		"name":         n.Name,
		"money":        n.Money,
		"trade_status": n.TradeStatus,
		"param":        n.Param,
	}
}
