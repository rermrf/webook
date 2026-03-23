// 第三方应用服务器示例
// 演示如何接入 webook 平台的积分支付系统
//
// 前置条件:
// 1. 启动 openapi 服务 (管理应用注册)
// 2. 启动 credit 服务 (积分支付)
// 3. 在 openapi 注册应用并通过审核
//
// 运行: go run main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"webook/credit/example/thirdparty"
)

// ========== 配置 ==========

const (
	// 第三方应用服务器地址
	ServerAddr = ":9000"

	// webook credit 服务地址
	CreditServiceURL = "http://localhost:8102"

	// 商户配置 (从 openapi 注册获取)
	MerchantPID = "app_demo_001"       // 应用ID
	MerchantKey = "your_api_secret_here" // API密钥
)

// ========== 订单存储 (示例用内存存储) ==========

type Order struct {
	OrderNo     string
	ProductName string
	Amount      int64  // 积分数量
	Uid         int64  // 用户ID
	Status      string // pending, paid, failed
	TradeNo     string // 平台订单号
	Param       string // 自定义参数
	CreatedAt   time.Time
	PaidAt      time.Time
}

var (
	orders   = make(map[string]*Order)
	ordersMu sync.RWMutex
)

// ========== SDK 实例 ==========

var sdk *thirdparty.CreditPaySDK

func init() {
	sdk = thirdparty.NewCreditPaySDK(CreditServiceURL, MerchantPID, MerchantKey)
}

// ========== 主函数 ==========

func main() {
	r := gin.Default()

	// 商品页面
	r.GET("/", indexPage)

	// 创建订单并发起支付
	r.POST("/order/create", createOrder)

	// 查询订单状态
	r.GET("/order/:orderNo", queryOrder)

	// 异步通知接收 (credit 服务回调)
	r.POST("/pay/notify", handleNotify)

	// 同步跳转页面 (用户支付完成后跳转)
	r.GET("/pay/return", handleReturn)

	log.Printf("第三方应用服务器启动: http://localhost%s", ServerAddr)
	log.Printf("Credit 服务地址: %s", CreditServiceURL)
	r.Run(ServerAddr)
}

// ========== 页面处理 ==========

// indexPage 商品展示页面
func indexPage(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>第三方商城 - 积分支付示例</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .product { border: 1px solid #ddd; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .price { color: #ff6b00; font-size: 24px; font-weight: bold; }
        .btn { background: #ff6b00; color: #fff; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .btn:hover { background: #e65c00; }
        input { padding: 8px; margin: 5px 0; width: 200px; }
    </style>
</head>
<body>
    <h1>🛒 第三方商城</h1>
    <p>演示接入 webook 积分支付</p>

    <div class="product">
        <h3>VIP会员月卡</h3>
        <p>享受一个月VIP特权</p>
        <p class="price">100 积分</p>
        <form action="/order/create" method="POST">
            <input type="hidden" name="product" value="VIP会员月卡">
            <input type="hidden" name="amount" value="100">
            <p><input type="number" name="uid" placeholder="输入用户ID" required></p>
            <button type="submit" class="btn">立即购买</button>
        </form>
    </div>

    <div class="product">
        <h3>VIP会员年卡</h3>
        <p>享受一年VIP特权，超值优惠</p>
        <p class="price">1000 积分</p>
        <form action="/order/create" method="POST">
            <input type="hidden" name="product" value="VIP会员年卡">
            <input type="hidden" name="amount" value="1000">
            <p><input type="number" name="uid" placeholder="输入用户ID" required></p>
            <button type="submit" class="btn">立即购买</button>
        </form>
    </div>

    <hr>
    <h3>查询订单</h3>
    <form action="/order/query" method="GET" onsubmit="window.location='/order/'+document.getElementById('orderNo').value;return false;">
        <input type="text" id="orderNo" placeholder="输入订单号">
        <button type="submit" class="btn">查询</button>
    </form>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ========== 订单处理 ==========

// createOrder 创建订单并跳转支付
func createOrder(c *gin.Context) {
	product := c.PostForm("product")
	amountStr := c.PostForm("amount")
	uidStr := c.PostForm("uid")

	amount, _ := strconv.ParseInt(amountStr, 10, 64)
	uid, _ := strconv.ParseInt(uidStr, 10, 64)

	if product == "" || amount <= 0 || uid <= 0 {
		c.String(http.StatusBadRequest, "参数错误")
		return
	}

	// 生成订单号
	orderNo := thirdparty.GenerateOutTradeNo("ORD")

	// 保存订单
	order := &Order{
		OrderNo:     orderNo,
		ProductName: product,
		Amount:      amount,
		Uid:         uid,
		Status:      "pending",
		Param:       fmt.Sprintf("product:%s", product),
		CreatedAt:   time.Now(),
	}
	ordersMu.Lock()
	orders[orderNo] = order
	ordersMu.Unlock()

	// 生成支付URL
	payURL, err := sdk.CreatePayURL(thirdparty.PayRequest{
		OutTradeNo: orderNo,
		Name:       product,
		Money:      amount,
		Uid:        uid,
		NotifyURL:  fmt.Sprintf("http://localhost%s/pay/notify", ServerAddr),
		ReturnURL:  fmt.Sprintf("http://localhost%s/pay/return", ServerAddr),
		Param:      order.Param,
	})
	if err != nil {
		c.String(http.StatusBadRequest, "创建支付失败: %v", err)
		return
	}

	log.Printf("[创建订单] orderNo=%s, product=%s, amount=%d, uid=%d", orderNo, product, amount, uid)

	// 重定向到支付页面
	c.Redirect(http.StatusFound, payURL)
}

// queryOrder 查询订单状态
func queryOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	// 查询本地订单
	ordersMu.RLock()
	order, exists := orders[orderNo]
	ordersMu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	// 同时查询平台订单状态
	platformStatus := ""
	if result, err := sdk.QueryOrder(orderNo); err == nil {
		platformStatus = result.TradeStatus
	}

	c.JSON(http.StatusOK, gin.H{
		"order_no":        order.OrderNo,
		"product":         order.ProductName,
		"amount":          order.Amount,
		"uid":             order.Uid,
		"status":          order.Status,
		"trade_no":        order.TradeNo,
		"platform_status": platformStatus,
		"created_at":      order.CreatedAt.Format("2006-01-02 15:04:05"),
		"paid_at":         order.PaidAt.Format("2006-01-02 15:04:05"),
	})
}

// ========== 支付回调处理 ==========

// handleNotify 处理异步通知
func handleNotify(c *gin.Context) {
	var params thirdparty.NotifyParams
	if err := c.ShouldBind(&params); err != nil {
		log.Printf("[异步通知] 参数解析失败: %v", err)
		c.String(http.StatusOK, "fail")
		return
	}

	log.Printf("[异步通知] 收到通知: orderNo=%s, tradeNo=%s, status=%s, money=%s",
		params.OutTradeNo, params.TradeNo, params.TradeStatus, params.Money)

	// 验证签名
	if !sdk.VerifyNotify(params) {
		log.Printf("[异步通知] 签名验证失败")
		c.String(http.StatusOK, "fail")
		return
	}

	// 验证订单
	ordersMu.Lock()
	order, exists := orders[params.OutTradeNo]
	if !exists {
		ordersMu.Unlock()
		log.Printf("[异步通知] 订单不存在: %s", params.OutTradeNo)
		c.String(http.StatusOK, "fail")
		return
	}

	// 验证金额
	if params.GetMoney() != order.Amount {
		ordersMu.Unlock()
		log.Printf("[异步通知] 金额不匹配: expected=%d, got=%d", order.Amount, params.GetMoney())
		c.String(http.StatusOK, "fail")
		return
	}

	// 更新订单状态
	if params.IsSuccess() && order.Status == "pending" {
		order.Status = "paid"
		order.TradeNo = params.TradeNo
		order.PaidAt = time.Now()
		log.Printf("[异步通知] 订单支付成功: %s", params.OutTradeNo)

		// TODO: 这里执行发货逻辑
		// deliverProduct(order)
	}
	ordersMu.Unlock()

	// 返回 success 表示通知处理成功
	c.String(http.StatusOK, "success")
}

// handleReturn 处理同步跳转
func handleReturn(c *gin.Context) {
	orderNo := c.Query("out_trade_no")
	tradeStatus := c.Query("trade_status")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>支付结果</title>
    <style>
        body { font-family: sans-serif; max-width: 400px; margin: 100px auto; text-align: center; }
        .success { color: #52c41a; }
        .pending { color: #faad14; }
        a { color: #1890ff; }
    </style>
</head>
<body>
    <h1 class="%s">%s</h1>
    <p>订单号: %s</p>
    <p><a href="/">返回首页</a> | <a href="/order/%s">查看订单</a></p>
</body>
</html>`,
		func() string {
			if tradeStatus == thirdparty.TradeStatusSuccess {
				return "success"
			}
			return "pending"
		}(),
		func() string {
			if tradeStatus == thirdparty.TradeStatusSuccess {
				return "✅ 支付成功"
			}
			return "⏳ 支付处理中"
		}(),
		orderNo,
		orderNo,
	)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
