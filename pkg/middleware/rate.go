package middleware

import (
	"github.com/ennismar/go-helper/pkg/resp"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"github.com/ulule/limiter/v3/drivers/store/redis"
	"net/http"
	"time"
)

func Rate(options ...func(*RateOptions)) gin.HandlerFunc {
	ops := getRateOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	rate := limiter.Rate{
		Period: time.Second,
		Limit:  ops.maxLimit,
	}

	var store limiter.Store
	var err error
	if ops.redis != nil {
		store, err = redis.NewStore(ops.redis)
		if err != nil {
			panic(err)
		}
	} else {
		store = memory.NewStore()
	}

	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))

	return mgin.NewMiddleware(instance, mgin.WithLimitReachedHandler(func(c *gin.Context) {
		rp := resp.GetFailWithCodeAndMsg(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		c.JSON(http.StatusTooManyRequests, rp)
	}))
}
