package middleware

import (
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/gin-gonic/gin"
)

func SecurityHeader(c *gin.Context) {
	ctx := tracing.RealCtx(c)
	_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "SecurityHeader"))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("X-Frame-Options", "deny")
	span.End()
	c.Next()
}
