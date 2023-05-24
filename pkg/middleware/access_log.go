package middleware

import (
	"bytes"
	"fmt"
	"github.com/ennismar/go-helper/pkg/constant"
	"github.com/ennismar/go-helper/pkg/log"
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"time"
)

type accessWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w accessWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AccessLog(options ...func(*AccessLogOptions)) gin.HandlerFunc {
	ops := getAccessLogOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		startTime := time.Now()

		w := &accessWriter{
			body:           bytes.NewBuffer(nil),
			ResponseWriter: c.Writer,
		}
		c.Writer = w

		getBody(c)

		c.Next()
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "AccessLog"))
		defer span.End()

		endTime := time.Now()

		// calc request exec time
		execTime := endTime.Sub(startTime).String()

		reqMethod := c.Request.Method
		reqPath := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		detail := make(map[string]interface{})
		if ops.detail {
			detail = getRequestDetail(c)
			span.SetAttributes(
				attribute.String(constant.MiddlewareParamsRespLogKey, detail[constant.MiddlewareParamsRespLogKey].(string)),
			)
		}

		detail[constant.MiddlewareAccessLogIpLogKey] = clientIP

		l := log.WithContext(c).WithFields(detail)

		if reqMethod == "OPTIONS" || reqPath == fmt.Sprintf("/%s/ping", ops.urlPrefix) {
			l.Debug(
				"%s %s %d %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
			)
		} else {
			l.Info(
				"%s %s %d %s",
				reqMethod,
				reqPath,
				statusCode,
				execTime,
			)
		}
	}
}
