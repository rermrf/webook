// Package thirdparty 第三方应用接入积分支付 SDK
package thirdparty

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CreditPaySDK 积分支付 SDK
type CreditPaySDK struct {
	BaseURL    string        // credit 服务地址
	PID        string        // 商户ID (app_id)
	Key        string        // 商户密钥 (api_secret)
	HTTPClient *http.Client  // HTTP 客户端
}

// NewCreditPaySDK 创建 SDK 实例
func NewCreditPaySDK(baseURL, pid, key string) *CreditPaySDK {
	return &CreditPaySDK{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		PID:     pid,
		Key:     key,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ========== 发起支付 ==========

// PayRequest 支付请求参数
type PayRequest struct {
	OutTradeNo string // 商户订单号（必填，唯一）
	Name       string // 商品名称（必填）
	Money      int64  // 积分数量（必填）
	Uid        int64  // 支付用户ID（必填）
	NotifyURL  string // 异步通知地址（必填）
	ReturnURL  string // 同步跳转地址（可选）
	Param      string // 自定义参数（可选，原样返回）
}

// Validate 验证请求参数
func (r *PayRequest) Validate() error {
	if r.OutTradeNo == "" {
		return errors.New("out_trade_no is required")
	}
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Money <= 0 {
		return errors.New("money must be positive")
	}
	if r.Uid <= 0 {
		return errors.New("uid must be positive")
	}
	if r.NotifyURL == "" {
		return errors.New("notify_url is required")
	}
	return nil
}

// CreatePayURL 生成支付URL，将用户重定向到此URL完成支付
func (sdk *CreditPaySDK) CreatePayURL(req PayRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	params := map[string]string{
		"pid":          sdk.PID,
		"type":         "credit",
		"out_trade_no": req.OutTradeNo,
		"notify_url":   req.NotifyURL,
		"return_url":   req.ReturnURL,
		"name":         req.Name,
		"money":        strconv.FormatInt(req.Money, 10),
		"param":        req.Param,
		"uid":          strconv.FormatInt(req.Uid, 10),
	}
	params["sign"] = sdk.Sign(params)
	params["sign_type"] = "MD5"

	u, _ := url.Parse(sdk.BaseURL + "/mapi/submit.php")
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ========== 查询订单 ==========

// QueryResponse 查询响应
type QueryResponse struct {
	Code        int    `json:"code"`
	Msg         string `json:"msg"`
	TradeNo     string `json:"trade_no"`
	OutTradeNo  string `json:"out_trade_no"`
	TradeStatus string `json:"trade_status"` // WAIT_PAY, TRADE_SUCCESS, TRADE_CLOSED
	Money       string `json:"money"`
	Name        string `json:"name"`
	Param       string `json:"param"`
	Addtime     string `json:"addtime"`
	Endtime     string `json:"endtime"`
}

// IsSuccess 判断订单是否支付成功
func (r *QueryResponse) IsSuccess() bool {
	return r.TradeStatus == "TRADE_SUCCESS"
}

// IsPending 判断订单是否待支付
func (r *QueryResponse) IsPending() bool {
	return r.TradeStatus == "WAIT_PAY"
}

// QueryOrder 查询订单状态
func (sdk *CreditPaySDK) QueryOrder(outTradeNo string) (*QueryResponse, error) {
	if outTradeNo == "" {
		return nil, errors.New("out_trade_no is required")
	}

	params := map[string]string{
		"pid":          sdk.PID,
		"out_trade_no": outTradeNo,
	}
	params["sign"] = sdk.Sign(params)
	params["sign_type"] = "MD5"

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	resp, err := sdk.HTTPClient.PostForm(sdk.BaseURL+"/mapi/query.php", form)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != 1 {
		return nil, fmt.Errorf("query failed: %s", result.Msg)
	}

	return &result, nil
}

// QueryByTradeNo 根据平台订单号查询
func (sdk *CreditPaySDK) QueryByTradeNo(tradeNo string) (*QueryResponse, error) {
	if tradeNo == "" {
		return nil, errors.New("trade_no is required")
	}

	params := map[string]string{
		"pid":      sdk.PID,
		"trade_no": tradeNo,
	}
	params["sign"] = sdk.Sign(params)
	params["sign_type"] = "MD5"

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	resp, err := sdk.HTTPClient.PostForm(sdk.BaseURL+"/mapi/query.php", form)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result QueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.Code != 1 {
		return nil, fmt.Errorf("query failed: %s", result.Msg)
	}

	return &result, nil
}

// ========== 异步通知处理 ==========

// NotifyParams 异步通知参数
type NotifyParams struct {
	PID         string `form:"pid" json:"pid"`
	TradeNo     string `form:"trade_no" json:"trade_no"`
	OutTradeNo  string `form:"out_trade_no" json:"out_trade_no"`
	Type        string `form:"type" json:"type"`
	Name        string `form:"name" json:"name"`
	Money       string `form:"money" json:"money"`
	TradeStatus string `form:"trade_status" json:"trade_status"`
	Param       string `form:"param" json:"param"`
	Sign        string `form:"sign" json:"sign"`
	SignType    string `form:"sign_type" json:"sign_type"`
}

// IsSuccess 判断是否支付成功
func (n *NotifyParams) IsSuccess() bool {
	return n.TradeStatus == "TRADE_SUCCESS"
}

// GetMoney 获取金额（int64）
func (n *NotifyParams) GetMoney() int64 {
	money, _ := strconv.ParseInt(n.Money, 10, 64)
	return money
}

// VerifyNotify 验证异步通知签名
func (sdk *CreditPaySDK) VerifyNotify(params NotifyParams) bool {
	m := map[string]string{
		"pid":          params.PID,
		"trade_no":     params.TradeNo,
		"out_trade_no": params.OutTradeNo,
		"type":         params.Type,
		"name":         params.Name,
		"money":        params.Money,
		"trade_status": params.TradeStatus,
		"param":        params.Param,
	}
	expected := sdk.Sign(m)
	return strings.EqualFold(expected, params.Sign)
}

// ========== 签名方法 ==========

// Sign 生成 MD5 签名
func (sdk *CreditPaySDK) Sign(params map[string]string) string {
	// 过滤空值和签名字段
	filtered := make(map[string]string)
	for k, v := range params {
		if v != "" && k != "sign" && k != "sign_type" {
			filtered[k] = v
		}
	}
	// 按 key 排序
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
	buf.WriteString(sdk.Key)
	// MD5
	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}

// ========== 工具方法 ==========

// GenerateOutTradeNo 生成商户订单号
func GenerateOutTradeNo(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
}

// TradeStatus 交易状态常量
const (
	TradeStatusWaitPay = "WAIT_PAY"      // 待支付
	TradeStatusSuccess = "TRADE_SUCCESS" // 支付成功
	TradeStatusClosed  = "TRADE_CLOSED"  // 已关闭
)
