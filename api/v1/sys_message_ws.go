package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/middleware"
	"github.com/piupuer/go-helper/pkg/query"
	"net/http"
)

// message websocket

func NewMessageHub(options ...func(*Options)) *query.MessageHub {
	ops := ParseOptions(options...)
	if ops.binlog {
		rd := query.NewRedis(ops.binlogOps...)
		ops.messageHubOps = append(ops.messageHubOps, query.WithMessageHubRedis(&rd))
	}
	my := query.NewMySql(ops.dbOps...)
	ops.messageHubOps = append(
		ops.messageHubOps,
		query.WithMessageHubDbNoTx(&my),
		query.WithMessageHubFindUserByIds(ops.findUserByIds),
	)
	return query.NewMessageHub(ops.messageHubOps...)
}

func MessageWs(hub *query.MessageHub, options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	if ops.getCurrentUser == nil {
		panic("getCurrentUser is empty")
	}
	return func(c *gin.Context) {
		h := make(http.Header)
		conn, err := middleware.WsUpgrader.Upgrade(c.Writer, c.Request, h)
		if err != nil {
			log.WithContext(c).WithError(err).Error("upgrade websocket failed")
			return
		}

		u := ops.getCurrentUser(c)
		hub.MessageWs(c, conn, c.Request.Header.Get("Sec-WebSocket-Key"), u, c.ClientIP())
	}
}
