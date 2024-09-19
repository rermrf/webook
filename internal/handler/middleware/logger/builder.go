package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"sync/atomic"
	"time"
)

// MiddlewareBuilder 注意点：
// 1. 小心日志内容过多。URL 可能很长，请求体，响应体都可能很大，需要考虑是不是完全输出到日志里面
// 2. 考虑 1 的问题，以及用户可能换用不同的日志框架，所以有足够的灵活性
// 3. 考虑动态开关，结合配置文件，要小心并发安全
type MiddlewareBuilder struct {
	// 控制请求体是否需要打印
	allowReqBody atomic.Bool
	// 控制响应体是否需要打印
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc:   fn,
		allowReqBody: atomic.Bool{},
	}
}

func (b *MiddlewareBuilder) AllowReqBody(ok bool) *MiddlewareBuilder {
	b.allowReqBody.Store(ok)
	return b
}

func (b *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	b.allowRespBody = true
	return b
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			// url 本身可能很长
			Url: url,
		}
		if b.allowReqBody.Load() && ctx.Request.Body != nil {
			// Body 读完就没有了，一次性操作
			body, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

			if len(body) > 1024 {
				body = body[:1024]
			}
			// 这其实是一个很消耗 CPU 和内存的操作
			// 因为会引起复制
			al.ReqBody = string(body)
		}

		if b.allowRespBody {
			ctx.Writer = &responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			al.Duration = time.Since(start).String()
			b.loggerFunc(ctx, al)
		}()

		// 执行到业务逻辑
		ctx.Next()

	}
}

// 组合装饰器模式，装饰部分方法
type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.al.RespBody = string(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.al.RespBody = s
	return w.ResponseWriter.WriteString(s)
}

type AccessLog struct {
	// HTTP 请求的方法
	Method string
	// URL 整个请求 URL
	Url      string
	Duration string
	ReqBody  string
	RespBody string
	status   int
}
