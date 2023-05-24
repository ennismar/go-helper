package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	"github.com/piupuer/go-helper/pkg/constant"
	"github.com/piupuer/go-helper/pkg/tracing"
	"github.com/piupuer/go-helper/pkg/utils"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	logCache = make([]OperationRecord, 0)
	logLock  sync.Mutex
)

type OperationApi struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Desc   string `json:"desc"`
}

type OperationRecord struct {
	CreatedAt  carbon.DateTime `json:"createdAt"`
	ApiDesc    string          `json:"apiDesc"`
	Path       string          `json:"path"`
	Method     string          `json:"method"`
	Header     string          `json:"header"`
	Body       string          `json:"body"`
	Params     string          `json:"params"`
	Resp       string          `json:"resp"`
	Status     int             `json:"status"`
	Username   string          `json:"username"`
	RoleName   string          `json:"roleName"`
	Ip         string          `json:"ip"`
	IpLocation string          `json:"ipLocation"`
	Latency    time.Duration   `json:"latency"`
	UserAgent  string          `json:"userAgent"`
}

func OperationLog(options ...func(*OperationLogOptions)) gin.HandlerFunc {
	ops := getOperationLogOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return func(c *gin.Context) {
		startTime := carbon.Now()
		reqBody := getBody(c)
		// find request params
		reqParams := c.Request.URL.Query()
		defer func() {
			ctx := tracing.RealCtx(c)
			_, span := tracer.Start(ctx, tracing.Name(tracing.Middleware, "OperationLog"))
			defer span.End()
			if ops.skipGetOrOptionsMethod {
				// skip GET/OPTIONS
				if c.Request.Method == http.MethodGet ||
					c.Request.Method == http.MethodOptions {
					return
				}
			}
			// custom skip path
			if ops.findSkipPath != nil {
				list := ops.findSkipPath(c)
				for _, item := range list {
					if strings.Contains(c.Request.URL.Path, item) {
						return
					}
				}
			}

			endTime := carbon.Now()

			contentType := c.Request.Header.Get("Content-Type")
			// multipart/form-data
			if strings.Contains(contentType, "multipart/form-data") {
				contentTypeArr := strings.Split(contentType, "; ")
				if len(contentTypeArr) == 2 {
					// read boundary
					boundary := strings.TrimPrefix(contentTypeArr[1], "boundary=")
					b := strings.NewReader(reqBody)
					r := multipart.NewReader(b, boundary)
					f, _ := r.ReadForm(ops.singleFileMaxSize << 20)
					defer f.RemoveAll()
					params := make(map[string]string, 0)
					for key, val := range f.Value {
						// get first value
						if len(val) > 0 {
							params[key] = val[0]
						}
					}
					params["content-type"] = "multipart/form-data"
					params["file"] = "binary data ignored"
					// save data by json format
					reqBody = utils.Struct2Json(params)
				}
			}
			// read header
			header := make(map[string]string, 0)
			for k, v := range c.Request.Header {
				header[k] = strings.Join(v, " | ")
			}
			record := OperationRecord{
				Ip:        c.ClientIP(),
				Method:    c.Request.Method,
				Path:      strings.TrimPrefix(c.Request.URL.Path, "/"+ops.urlPrefix),
				Header:    utils.Struct2Json(header),
				Body:      reqBody,
				Params:    utils.Struct2Json(reqParams),
				Latency:   endTime.Carbon2Time().Sub(startTime.Carbon2Time()),
				UserAgent: c.Request.UserAgent(),
			}
			record.CreatedAt = carbon.DateTime{
				Carbon: endTime,
			}

			record.Username = constant.MiddlewareOperationLogNotLogin
			record.RoleName = constant.MiddlewareOperationLogNotLogin
			// get current user
			u := ops.getCurrentUser(c)
			if u.RoleName != "" {
				record.Username = u.Username
				record.RoleName = u.RoleName
			}

			record.ApiDesc = getApiDesc(c, record.Method, record.Path, *ops)
			// get ip location
			record.IpLocation = utils.GetIpRealLocation(record.Ip, ops.realIpKey)

			record.Status = c.Writer.Status()

			// record.Resp = response

			// delay to update to db
			logLock.Lock()
			logCache = append(logCache, record)
			if len(logCache) >= ops.maxCountBeforeSave {
				list := make([]OperationRecord, len(logCache))
				copy(list, logCache)
				go ops.save(c, list)
				logCache = make([]OperationRecord, 0)
			}
			logLock.Unlock()
		}()
		c.Next()
	}
}

func getApiDesc(c *gin.Context, method, path string, ops OperationLogOptions) string {
	desc := "no desc"
	if ops.redis != nil {
		oldCache, _ := ops.redis.HGet(c, ops.cachePrefix, fmt.Sprintf("%s_%s", method, path)).Result()
		if oldCache != "" {
			return oldCache
		}
	}
	apis := ops.findApi(c)
	for _, api := range apis {
		if api.Method == method && api.Path == path {
			desc = api.Desc
			break
		}
	}

	if ops.redis != nil {
		pipe := ops.redis.Pipeline()
		for _, api := range apis {
			pipe.HSet(c, ops.cachePrefix, fmt.Sprintf("%s_%s", api.Method, api.Path), api.Desc)
		}
		pipe.Exec(c)
	}
	return desc
}
