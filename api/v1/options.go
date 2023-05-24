package v1

import (
	"context"
	"github.com/ennismar/go-helper/ms"
	"github.com/ennismar/go-helper/pkg/delay"
	"github.com/ennismar/go-helper/pkg/oss"
	"github.com/ennismar/go-helper/pkg/query"
	"github.com/ennismar/go-helper/pkg/req"
	"github.com/ennismar/go-helper/pkg/resp"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type Options struct {
	binlog                      bool
	binlogOps                   []func(options *query.RedisOptions)
	dbOps                       []func(options *query.MysqlOptions)
	exportOps                   []func(options *delay.ExportOptions)
	redis                       redis.UniversalClient
	cachePrefix                 string
	operationAllowedToDelete    bool
	getCurrentUser              func(c *gin.Context) ms.User
	findRoleKeywordByRoleIds    func(c *gin.Context, roleIds []uint) []string
	findRoleByIds               func(c *gin.Context, roleIds []uint) []ms.Role
	findUserByIds               func(c *gin.Context, userIds []uint) []ms.User
	beforeResetUserPwd          func(c *gin.Context, r *req.ResetUserPwd) error
	getUserLoginStatus          func(c *gin.Context, r *req.UserStatus) error
	getFsmLogSubmitterDetail    func(c *gin.Context, detail req.FsmLogSubmitterDetail) []resp.FsmLogSubmitterDetail
	updateFsmLogSubmitterDetail func(c *gin.Context, detail req.UpdateFsmLogSubmitterDetail) error
	uploadSaveDir               string
	uploadSingleMaxSize         int64
	uploadMergeConcurrentCount  int
	uploadMinio                 *oss.MinioOss
	uploadMinioBucket           string
	MessageHub                  bool
	messageHubOps               []func(options *query.MessageHubOptions)
}

func WithBinlog(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).binlog = flag
	}
}

func WithBinlogOps(ops ...func(options *query.RedisOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).binlogOps = append(getOptionsOrSetDefault(options).binlogOps, ops...)
	}
}

func WithDbOps(ops ...func(options *query.MysqlOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).dbOps = append(getOptionsOrSetDefault(options).dbOps, ops...)
	}
}

func WithExportOps(ops ...func(options *delay.ExportOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).exportOps = append(getOptionsOrSetDefault(options).exportOps, ops...)
	}
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithCachePrefix(prefix string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).cachePrefix = prefix
	}
}

func WithOperationAllowedToDelete(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).operationAllowedToDelete = flag
	}
}

func WithGetCurrentUser(fun func(c *gin.Context) ms.User) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).getCurrentUser = fun
		}
	}
}

func WithFindRoleKeywordByRoleIds(fun func(c *gin.Context, roleIds []uint) []string) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).findRoleKeywordByRoleIds = fun
		}
	}
}

func WithFindRoleByIds(fun func(c *gin.Context, roleIds []uint) []ms.Role) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).findRoleByIds = fun
		}
	}
}

func WithFindUserByIds(fun func(c *gin.Context, userIds []uint) []ms.User) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).findUserByIds = fun
		}
	}
}

func WithBeforeResetUserPwd(fun func(c *gin.Context, r *req.ResetUserPwd) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).beforeResetUserPwd = fun
		}
	}
}

func WithGetUserLoginStatus(fun func(c *gin.Context, r *req.UserStatus) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).getUserLoginStatus = fun
		}
	}
}

func WithFsmGetFsmLogSubmitterDetail(fun func(c *gin.Context, detail req.FsmLogSubmitterDetail) []resp.FsmLogSubmitterDetail) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).getFsmLogSubmitterDetail = fun
		}
	}
}

func WithFsmUpdateFsmLogSubmitterDetail(fun func(c *gin.Context, detail req.UpdateFsmLogSubmitterDetail) error) func(*Options) {
	return func(options *Options) {
		if fun != nil {
			getOptionsOrSetDefault(options).updateFsmLogSubmitterDetail = fun
		}
	}
}

func WithUploadSaveDir(dir string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).uploadSaveDir = dir
	}
}

func WithUploadSingleMaxSize(size int64) func(*Options) {
	return func(options *Options) {
		if size > 0 {
			getOptionsOrSetDefault(options).uploadSingleMaxSize = size
		}
	}
}

func WithUploadMergeConcurrentCount(count int) func(*Options) {
	return func(options *Options) {
		if count > 1 {
			getOptionsOrSetDefault(options).uploadMergeConcurrentCount = count
		}
	}
}

func WithUploadMinio(minio *oss.MinioOss) func(*Options) {
	return func(options *Options) {
		if minio != nil {
			getOptionsOrSetDefault(options).uploadMinio = minio
		}
	}
}

func WithUploadMinioBucket(bucket string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).uploadMinioBucket = bucket
	}
}

func WithMessageHub(flag bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).MessageHub = flag
	}
}

func WithMessageHubOps(ops ...func(options *query.MessageHubOptions)) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).messageHubOps = append(getOptionsOrSetDefault(options).messageHubOps, ops...)
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			binlog:                     false,
			cachePrefix:                "v1_cache",
			operationAllowedToDelete:   true,
			uploadSaveDir:              "upload",
			uploadSingleMaxSize:        32,
			uploadMergeConcurrentCount: 10,
			MessageHub:                 true,
		}
	}
	return options
}

func ParseOptions(options ...func(*Options)) *Options {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// check ops
	if ops.binlog {
		if ops.redis != nil {
			ops.binlogOps = append(ops.binlogOps, query.WithRedisClient(ops.redis))
		}
		query.NewRedis(ops.binlogOps...)
	}
	if ops.redis != nil {
		ops.dbOps = append(ops.dbOps, query.WithMysqlRedis(ops.redis))
	}
	query.NewMySql(ops.dbOps...)
	return ops
}

func (ops *Options) addCtx(ctx context.Context) {
	if ops.binlog {
		ops.binlogOps = append(ops.binlogOps, query.WithRedisCtx(ctx))
	}
	ops.dbOps = append(ops.dbOps, query.WithMysqlCtx(ctx))
	ops.exportOps = append(ops.exportOps, delay.WithExportCtx(ctx))
}
