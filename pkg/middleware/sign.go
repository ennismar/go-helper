package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/ennismar/go-helper/pkg/constant"
	"github.com/ennismar/go-helper/pkg/log"
	"github.com/ennismar/go-helper/pkg/resp"
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/ennismar/go-helper/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	"net/http"
	"regexp"
	"strings"
)

func Sign(options ...func(*SignOptions)) gin.HandlerFunc {
	ops := getSignOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.getSignUser == nil {
		panic("getSignUser is empty")
	}
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "Sign"))
		var pass bool
		defer func() {
			if !pass {
				span.End()
			}
		}()
		if ops.findSkipPath != nil {
			list := ops.findSkipPath(c)
			for _, item := range list {
				if strings.Contains(c.Request.URL.Path, item) {
					span.End()
					pass = true
					c.Next()
					return
				}
				re, err := regexp.Compile(item)
				if err != nil {
					continue
				}
				if re.MatchString(c.Request.URL.Path) {
					span.End()
					pass = true
					c.Next()
					return
				}
			}
		}
		// token
		token := c.Request.Header.Get(ops.headerKey[0])
		if token == "" {
			log.WithContext(c).Warn(resp.InvalidSignTokenMsg)
			abort(c, resp.InvalidSignTokenMsg)
			return
		}
		list := strings.Split(token, ",")
		re := regexp.MustCompile(`"[\D\d].*"`)
		var appId, timestamp, signature string
		for _, item := range list {
			ms := re.FindAllString(item, -1)
			if len(ms) == 1 {
				if strings.HasPrefix(item, ops.headerKey[1]) {
					appId = strings.Trim(ms[0], `"`)
				} else if strings.HasPrefix(item, ops.headerKey[2]) {
					timestamp = strings.Trim(ms[0], `"`)
				} else if strings.HasPrefix(item, ops.headerKey[3]) {
					signature = strings.Trim(ms[0], `"`)
				}
			}
		}
		if appId == "" {
			log.WithContext(c).Warn(resp.InvalidSignIdMsg)
			abort(c, resp.InvalidSignIdMsg)
			return
		}
		if timestamp == "" {
			log.WithContext(c).Warn(resp.InvalidSignTimestampMsg)
			abort(c, resp.InvalidSignTimestampMsg)
			return
		}
		// compare timestamp
		now := carbon.Now()
		t := carbon.CreateFromTimestamp(utils.Str2Int64(timestamp))
		if t.AddDuration(ops.expire).Lt(now) {
			log.WithContext(c).Warn("%s: %s", resp.InvalidSignTimestampMsg, timestamp)
			abort(c, "%s: %s", resp.InvalidSignTimestampMsg, timestamp)
			return
		}
		// query user by app id
		u := ops.getSignUser(c, appId)
		if u.AppSecret == "" {
			log.WithContext(c).Warn("%s: %s", resp.IllegalSignIdMsg, appId)
			abort(c, "%s: %s", resp.IllegalSignIdMsg, appId)
			return
		}
		if u.Status == constant.Zero {
			log.WithContext(c).Warn("%s: %s", resp.UserDisabledMsg, appId)
			abort(c, "%s: %s", resp.UserDisabledMsg, appId)
			return
		}
		// scope
		reqMethod := c.Request.Method
		reqUri := c.Request.RequestURI
		if ops.checkScope {
			reqPath := c.Request.URL.Path
			exists := false
			for _, item := range u.Scopes {
				if item.Method == reqMethod && item.Path == reqPath {
					exists = true
					break
				}
			}
			if !exists {
				log.WithContext(c).Warn("%s: %s, %s", resp.InvalidSignScopeMsg, reqMethod, reqPath)
				abort(c, "%s: %s, %s", resp.InvalidSignScopeMsg, reqMethod, reqPath)
				return
			}
		}

		// verify signature
		if !verifySign(u.AppSecret, signature, reqMethod, reqUri, timestamp, getBody(c)) {
			log.WithContext(c).Warn("%s: %s", resp.IllegalSignTokenMsg, token)
			abort(c, "%s: %s", resp.IllegalSignTokenMsg, token)
			return
		}
		span.End()
		pass = true
		c.Next()
	}
}

func verifySign(secret, signature, method, uri, timestamp, body string) (flag bool) {
	b := bytes.NewBuffer(nil)
	b.WriteString(method)
	b.WriteString(constant.MiddlewareSignSeparator)
	b.WriteString(uri)
	b.WriteString(constant.MiddlewareSignSeparator)
	b.WriteString(timestamp)
	b.WriteString(constant.MiddlewareSignSeparator)
	b.WriteString(utils.JsonWithSort(body))
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(b.Bytes())
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	flag = digest == signature
	return
}

func abort(c *gin.Context, format interface{}, a ...interface{}) {
	rp := resp.GetFailWithMsg(format, a...)
	rp.RequestId, _, _ = tracing.GetId(c)
	c.JSON(http.StatusForbidden, rp)
	c.Abort()
}
