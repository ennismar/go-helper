package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"net/http"
)

func Transaction(options ...func(*TransactionOptions)) gin.HandlerFunc {
	ops := getTransactionOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.dbNoTx == nil {
		panic("dbNoTx is empty")
	}
	return func(c *gin.Context) {
		method := c.Request.Method
		noTransaction := false
		if method == "OPTIONS" || method == "GET" {
			// OPTIONS/GET skip transaction
			noTransaction = true
		}
		if funk.ContainsString(ops.forceTransactionPath, c.Request.URL.Path) {
			noTransaction = false
		}
		defer func() {
			ctx := tracing.RealCtx(c)
			_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "Transaction"))
			defer span.End()
			// get db transaction
			tx := getTx(c, *ops)
			if err := recover(); err != nil {
				if rp, ok := err.(resp.Resp); ok {
					if !noTransaction {
						if rp.Code == resp.Ok || c.GetBool(constant.MiddlewareTransactionForceCommitCtxKey) {
							// commit transaction
							tx.Commit()
						} else {
							// rollback transaction
							tx.Rollback()
						}
					}
					rp.RequestId, _, _ = tracing.GetId(c)
					c.JSON(http.StatusOK, rp)
					c.Abort()
					return
				}
				if !noTransaction {
					tx.Rollback()
				}
				// throw up exception
				panic(err)
			} else {
				if !noTransaction {
					tx.Commit()
				}
			}
			c.Abort()
		}()
		if !noTransaction {
			tx := ops.dbNoTx.Begin()
			c.Set(constant.MiddlewareTransactionTxCtxKey, tx)
		}
		c.Next()
	}
}

func getTx(c *gin.Context, ops TransactionOptions) *gorm.DB {
	tx := ops.dbNoTx
	txKey, exists := c.Get(constant.MiddlewareTransactionTxCtxKey)
	if exists {
		if item, ok := txKey.(*gorm.DB); ok {
			tx = item
		}
	}
	return tx
}
