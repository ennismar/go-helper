package query

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	"github.com/gorilla/websocket"
	"github.com/piupuer/go-helper/ms"
	"github.com/piupuer/go-helper/pkg/ch"
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/middleware"
	"github.com/piupuer/go-helper/pkg/req"
	"github.com/piupuer/go-helper/pkg/resp"
	"github.com/piupuer/go-helper/pkg/utils"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

// Message websocket

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Send heartbeat to peer with this period
	heartBeatPeriod = 10 * time.Second

	// Online notifications will not be sent repeatedly during the cycle
	lastActiveRegisterPeriod = 10 * time.Minute

	// Maximum number of heartbeat retries
	heartBeatMaxRetryCount = 3

	// Message Type
	// first number: request/response
	// second number: message type
	// third number: same message type's sort
	//
	// request message type(first number=1)
	MessageReqHeartBeat    string = "1-1-1"
	MessageReqPush         string = "1-2-1"
	MessageReqBatchRead    string = "1-2-2"
	MessageReqBatchDeleted string = "1-2-3"
	MessageReqAllRead      string = "1-2-4"
	MessageReqAllDeleted   string = "1-2-5"

	// response message type(first number=2)
	MessageRespHeartBeat string = "2-1-1"
	MessageRespNormal    string = "2-2-1"
	MessageRespUnRead    string = "2-3-1"
	MessageRespOnline    string = "2-4-1"
)

// var hub MessageHub

// MessageHub The message hub is used to maintain the entire message connection
type MessageHub struct {
	ops  MessageHubOptions
	lock sync.RWMutex
	// client user ids
	userIds []uint
	// client user active timestamp
	userLastActive map[uint]int64
	// MessageClients(key is websocket key)
	clients map[string]*MessageClient
	// The channel to refresh user message
	refreshUserMessage *ch.Ch
}

// MessageClient The message client is used to store connection information
type MessageClient struct {
	hub *MessageHub
	ctx context.Context
	// websocket key
	Key string
	// websocket connection
	Conn *websocket.Conn
	// current user
	User ms.User
	// request ip
	Ip string
	// The channel sends messages to user
	Send           *ch.Ch
	LastActiveTime carbon.Carbon
	RetryCount     uint
}

// MessageBroadcast The message broadcast is used to store users who need to broadcast
type MessageBroadcast struct {
	resp.MessageWs
	UserIds []uint `json:"-"`
}

func NewMessageHub(options ...func(*MessageHubOptions)) *MessageHub {
	ops := getMessageHubOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.dbNoTx == nil {
		panic("message hub dbNoTx is empty")
	}
	if ops.findUserByIds == nil {
		panic("message hub findUserByIds is empty")
	}
	hub := &MessageHub{
		ops: *ops,
	}
	hub.ops = *ops
	hub.clients = make(map[string]*MessageClient)
	hub.userLastActive = make(map[uint]int64)
	hub.refreshUserMessage = ch.NewCh()
	go hub.run()
	go hub.count()
	return hub
}

// MessageWs websocket handler
func (h *MessageHub) MessageWs(ctx *gin.Context, conn *websocket.Conn, key string, user ms.User, ip string) {
	client := &MessageClient{
		hub:  h,
		ctx:  ctx,
		Key:  key,
		Conn: conn,
		User: user,
		Ip:   ip,
		Send: ch.NewCh(),
	}

	go client.register()
	go client.receive()
	go client.send()
	go client.heartBeat()
}

func (h *MessageHub) SendToUserIds(userIds []uint, msg resp.MessageWs) {
	for _, client := range h.getClients() {
		// Notify the specified user
		if utils.ContainsUint(userIds, client.User.Id) {
			client.Send.SafeSend(msg)
		}
	}
}

func (h *MessageHub) run() {
	for {
		select {
		case data := <-h.refreshUserMessage.C:
			userIds := data.([]uint)
			users := h.ops.findUserByIds(&gin.Context{}, userIds)
			// sync users message
			h.ops.dbNoTx.SyncMessageByUserIds(users)
			for _, client := range h.getClients() {
				for _, id := range userIds {
					if client.User.Id == id {
						var total int64
						if h.ops.rd != nil {
							total = h.ops.rd.GetUnReadMessageCount(id)
						} else {
							total = h.ops.dbNoTx.GetUnReadMessageCount(id)
						}
						msg := resp.MessageWs{
							Type: MessageRespUnRead,
							Detail: resp.GetSuccessWithData(map[string]int64{
								"unReadCount": total,
							}),
						}
						client.Send.SafeSend(msg)
					}
				}
			}
		}
	}
}

// active connection count
func (h *MessageHub) count() {
	ticker := time.NewTicker(heartBeatPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			infos := make([]string, 0)
			for _, client := range h.getClients() {
				infos = append(infos, fmt.Sprintf("%d-%s", client.User.Id, client.Ip))
			}
			log.Debug("[message]active connection: %v", strings.Join(infos, ","))
		}
	}
}

func (h *MessageHub) getClients() map[string]*MessageClient {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.clients
}

// receive message handler
func (c *MessageClient) receive() {
	defer func() {
		c.close()
		if err := recover(); err != nil {
			log.WithContext(c.ctx).WithError(errors.Errorf("%v", err)).Error("[message][receiver][%s]connection may have been lost", c.Key)
		}
	}()
	for {
		_, msg, err := c.Conn.ReadMessage()

		// save active time
		c.LastActiveTime = carbon.Now()
		c.RetryCount = 0

		if err != nil {
			panic(err)
		}
		data := string(msg)
		if c.hub.ops.compress {
			// decompress data
			data = utils.DeCompressStrByZlib(string(msg))
		}
		log.WithContext(c.ctx).Debug("[message][receiver][%s]receive data success: %d, %s", c.Key, c.User.Id, data)
		var r req.MessageWs
		utils.Json2Struct(data, &r)
		switch r.Type {
		case MessageReqHeartBeat:
			if _, ok := r.Data.(float64); ok {
				c.Send.SafeSend(resp.MessageWs{
					Type:   MessageRespHeartBeat,
					Detail: resp.GetSuccess(),
				})
			}
		case MessageReqPush:
			var d req.PushMessage
			utils.Struct2StructByJson(r.Data, &d)
			err = req.ValidateWithErr(c.ctx, d, d.FieldTrans())
			detail := resp.GetSuccess()
			if err == nil {
				ops := middleware.ParseIdempotenceOptions(c.hub.ops.idempotenceOps...)
				if c.hub.ops.idempotence && !middleware.CheckIdempotenceToken(c.ctx, d.IdempotenceToken, *ops) {
					err = errors.Errorf(resp.IdempotenceTokenInvalidMsg)
				} else {
					d.FromUserId = c.User.Id
					err = c.hub.ops.dbNoTx.CreateMessage(&d)
				}
			}
			if err != nil {
				detail = resp.GetFailWithMsg(err)
			} else {
				c.hub.refreshUserMessage.SafeSend(c.hub.userIds)
			}
			c.Send.SafeSend(resp.MessageWs{
				Type:   MessageRespNormal,
				Detail: detail,
			})
		case MessageReqBatchRead:
			var d req.Ids
			utils.Struct2StructByJson(r.Data, &d)
			c.hub.ops.dbNoTx.BatchUpdateMessageRead(d.Uints())
			detail := resp.GetSuccess()
			c.hub.refreshUserMessage.SafeSend(c.hub.userIds)
			c.Send.SafeSend(resp.MessageWs{
				Type:   MessageRespNormal,
				Detail: detail,
			})
		case MessageReqBatchDeleted:
			var d req.Ids
			utils.Struct2StructByJson(r.Data, &d)
			c.hub.ops.dbNoTx.BatchUpdateMessageDeleted(d.Uints())
			detail := resp.GetSuccess()
			c.hub.refreshUserMessage.SafeSend(c.hub.userIds)
			c.Send.SafeSend(resp.MessageWs{
				Type:   MessageRespNormal,
				Detail: detail,
			})
		case MessageReqAllRead:
			c.hub.ops.dbNoTx.UpdateAllMessageRead(c.User.Id)
			detail := resp.GetSuccess()
			c.hub.refreshUserMessage.SafeSend(c.hub.userIds)
			c.Send.SafeSend(resp.MessageWs{
				Type:   MessageRespNormal,
				Detail: detail,
			})
		case MessageReqAllDeleted:
			c.hub.ops.dbNoTx.UpdateAllMessageDeleted(c.User.Id)
			detail := resp.GetSuccess()
			c.hub.refreshUserMessage.SafeSend(c.hub.userIds)
			c.Send.SafeSend(resp.MessageWs{
				Type:   MessageRespNormal,
				Detail: detail,
			})
		default:
			//
		}
	}
}

// send message handler
func (c *MessageClient) send() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
		if err := recover(); err != nil {
			log.WithContext(c.ctx).WithError(errors.Errorf("%v", err)).Error("[message][sender][%s]connection may have been lost", c.Key)
		}
	}()
	for {
		select {
		case msg, ok := <-c.Send.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// send failed
				c.writeMessage(websocket.CloseMessage, "closed")
				panic("connection closed")
			}

			if err := c.writeMessage(websocket.TextMessage, utils.Struct2Json(msg)); err != nil {
				panic(err)
			}
		// timeout
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.writeMessage(websocket.PingMessage, "ping"); err != nil {
				panic(err)
			}
		}
	}
}

func (c MessageClient) writeMessage(messageType int, data string) error {
	s := data
	if c.hub.ops.compress {
		// compress data
		s, _ = utils.CompressStrByZlib(data)
	}
	log.WithContext(c.ctx).Debug("[message][sender][%s] %v", c.Key, s)
	return c.Conn.WriteMessage(messageType, []byte(s))
}

// heartbeat handler
func (c *MessageClient) heartBeat() {
	ticker := time.NewTicker(heartBeatPeriod)
	defer func() {
		ticker.Stop()
		c.close()
		if err := recover(); err != nil {
			log.WithContext(c.ctx).WithError(errors.Errorf("%v", err)).Error("[message][heartbeat][%s]connection may have been lost", c.Key)
		}
	}()
	for {
		select {
		case <-ticker.C:
			last := time.Now().Sub(c.LastActiveTime.Carbon2Time())
			if c.RetryCount > heartBeatMaxRetryCount {
				panic(fmt.Sprintf("[message][heartbeat]retry sending heartbeat for %d times without response", c.RetryCount))
			}
			if last > heartBeatPeriod {
				c.Send.SafeSend(resp.MessageWs{
					Type:   MessageRespHeartBeat,
					Detail: resp.GetSuccessWithData(c.RetryCount),
				})
				c.RetryCount++
			} else {
				c.RetryCount = 0
			}
		}
	}
}

// user online handler
func (c *MessageClient) register() {
	c.hub.lock.Lock()
	defer c.hub.lock.Unlock()

	t := carbon.Now()
	active, ok := c.hub.userLastActive[c.User.Id]
	last := carbon.CreateFromTimestamp(active)
	c.hub.clients[c.Key] = c
	if !ok || last.AddDuration(lastActiveRegisterPeriod.String()).Lt(t) {
		if !utils.ContainsUint(c.hub.userIds, c.User.Id) {
			c.hub.userIds = append(c.hub.userIds, c.User.Id)
		}
		log.WithContext(c.ctx).Debug("[message][online][%s]%d-%s", c.Key, c.User.Id, c.Ip)
		go func() {
			c.hub.refreshUserMessage.SafeSend([]uint{c.User.Id})
		}()

		msg := resp.MessageWs{
			Type: MessageRespOnline,
			Detail: resp.GetSuccessWithData(map[string]interface{}{
				"user": c.User,
			}),
		}
		// Inform everyone except yourself
		go c.hub.SendToUserIds(utils.ContainsUintThenRemove(c.hub.userIds, c.User.Id), msg)

		c.hub.userLastActive[c.User.Id] = t.Timestamp()
	} else {
		c.hub.userLastActive[c.User.Id] = t.Timestamp()
	}
}

// user offline handler
func (c *MessageClient) close() {
	c.hub.lock.Lock()
	defer c.hub.lock.Unlock()

	if _, ok := c.hub.clients[c.Key]; ok {
		delete(c.hub.clients, c.Key)
		c.Send.SafeClose()
		log.WithContext(c.ctx).Debug("[message][offline][%s]%d-%s", c.Key, c.User.Id, c.Ip)
	}

	c.Conn.Close()
}
