# 第三方应用接入积分支付示例

本示例演示如何通过易支付协议接入 webook 平台的积分支付系统。

## 架构图

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   第三方应用     │     │  Credit 服务     │     │  OpenAPI 服务   │
│   (本示例)       │     │  (积分支付)      │     │  (应用管理)     │
│   :9000         │     │  :8102          │     │  :8103          │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         │  1. 创建支付URL        │                       │
         │ ─────────────────────>│                       │
         │                       │  2. 验证商户签名       │
         │                       │ ─────────────────────>│
         │                       │<─────────────────────│
         │  3. 重定向到支付页面   │                       │
         │<─────────────────────│                       │
         │                       │                       │
         │     用户确认支付       │                       │
         │                       │                       │
         │  4. 异步通知          │                       │
         │<─────────────────────│                       │
         │                       │                       │
         │  5. 返回 "success"    │                       │
         │ ─────────────────────>│                       │
         │                       │                       │
```

## 前置条件

### 1. 启动服务

```bash
# 启动 openapi 服务 (应用管理)
cd openapi && go run .

# 启动 credit 服务 (积分支付)
cd credit && go run .
```

### 2. 注册第三方应用

在 openapi 服务注册应用并获取凭证：

```bash
# 创建应用 (需要登录获取 JWT token)
curl -X POST http://localhost:8103/api/openapi/apps \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "第三方商城",
    "type": 2,
    "description": "演示积分支付接入",
    "scopes": ["write:credit"],
    "callback_url": "http://localhost:9000/pay/notify"
  }'

# 响应示例:
# {
#   "data": {
#     "app_id": "app_demo_001",
#     "app_secret": "your_api_secret_here",
#     "name": "第三方商城"
#   }
# }
```

### 3. 审核通过应用

```bash
# 管理员审核通过
curl -X POST http://localhost:8103/api/openapi/admin/apps/app_demo_001/audit \
  -H "Authorization: Bearer <admin_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}'
```

## 运行示例

### 1. 修改配置

编辑 `main.go` 中的配置：

```go
const (
    MerchantPID = "app_demo_001"         // 替换为你的 app_id
    MerchantKey = "your_api_secret_here" // 替换为你的 api_secret
)
```

### 2. 启动第三方应用

```bash
cd credit/example
go run main.go
```

### 3. 测试支付流程

1. 打开浏览器访问 http://localhost:9000
2. 输入用户ID，点击"立即购买"
3. 跳转到积分支付页面，确认支付
4. 支付成功后跳转回第三方应用

## SDK 使用说明

### 初始化

```go
import "webook/credit/example/thirdparty"

sdk := thirdparty.NewCreditPaySDK(
    "http://localhost:8102",  // credit 服务地址
    "app_demo_001",           // 商户ID
    "your_api_secret",        // 商户密钥
)
```

### 发起支付

```go
payURL, err := sdk.CreatePayURL(thirdparty.PayRequest{
    OutTradeNo: "ORDER_123456",           // 商户订单号
    Name:       "VIP会员",                 // 商品名称
    Money:      100,                       // 积分数量
    Uid:        12345,                     // 支付用户ID
    NotifyURL:  "https://your.com/notify", // 异步通知地址
    ReturnURL:  "https://your.com/return", // 同步跳转地址
    Param:      "custom_data",             // 自定义参数
})
if err != nil {
    // 处理参数错误
}

// 重定向用户到 payURL
http.Redirect(w, r, payURL, http.StatusFound)
```

### 处理异步通知

```go
func handleNotify(c *gin.Context) {
    var params thirdparty.NotifyParams
    c.ShouldBind(&params)

    // 1. 验证签名
    if !sdk.VerifyNotify(params) {
        c.String(200, "fail")
        return
    }

    // 2. 验证订单和金额
    // ...

    // 3. 处理业务逻辑
    if params.IsSuccess() {
        // 发货、开通服务等
    }

    // 4. 返回 success
    c.String(200, "success")
}
```

### 查询订单

```go
result, err := sdk.QueryOrder("ORDER_123456")
if err != nil {
    // 处理错误
}

if result.IsSuccess() {
    // 支付成功
} else if result.IsPending() {
    // 待支付
}

// 或使用常量
switch result.TradeStatus {
case thirdparty.TradeStatusWaitPay:
    // 待支付
case thirdparty.TradeStatusSuccess:
    // 支付成功
case thirdparty.TradeStatusClosed:
    // 已关闭
}
```

## 易支付接口规范

### 发起支付 `/mapi/submit.php`

| 参数 | 必填 | 说明 |
|------|------|------|
| pid | 是 | 商户ID |
| type | 否 | 支付类型，默认 credit |
| out_trade_no | 是 | 商户订单号 |
| notify_url | 是 | 异步通知地址 |
| return_url | 否 | 同步跳转地址 |
| name | 是 | 商品名称 |
| money | 是 | 积分数量 |
| uid | 是 | 支付用户ID |
| param | 否 | 自定义参数 |
| sign | 是 | MD5签名 |
| sign_type | 否 | 签名类型，默认 MD5 |

### 查询订单 `/mapi/query.php`

| 参数 | 必填 | 说明 |
|------|------|------|
| pid | 是 | 商户ID |
| out_trade_no | 二选一 | 商户订单号 |
| trade_no | 二选一 | 平台订单号 |
| sign | 是 | MD5签名 |

### 异步通知参数

| 参数 | 说明 |
|------|------|
| pid | 商户ID |
| trade_no | 平台订单号 |
| out_trade_no | 商户订单号 |
| type | 支付类型 |
| name | 商品名称 |
| money | 积分数量 |
| trade_status | 交易状态 |
| param | 自定义参数 |
| sign | MD5签名 |

### 签名规则

1. 过滤空值和 sign、sign_type 字段
2. 按参数名 ASCII 升序排序
3. 拼接成 `key1=value1&key2=value2` 格式
4. 末尾追加商户密钥
5. 计算 MD5 (小写)

```go
// 示例
params := "money=100&name=VIP&notify_url=http://...&out_trade_no=ORDER_123&pid=app_001&type=credit&uid=12345"
sign := md5(params + "your_secret")
```
