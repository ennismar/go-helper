package middleware

import (
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/gin-gonic/gin"
)

func RequestId(c *gin.Context) {
	ctx := tracing.RealCtx(c)
	_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "RequestId"))
	requestId, _, _ := tracing.GetId(c)
	if requestId == "" {
		c.Request = c.Request.WithContext(tracing.NewId(c))
	}
	span.End()
	c.Next()
}
