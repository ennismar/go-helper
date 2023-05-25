package delay

import (
	"context"
	"fmt"
	"github.com/ennismar/go-helper/pkg/constant"
	"github.com/ennismar/go-helper/pkg/utils"
	"gorm.io/gorm"
	"time"
)

type ExportOptions struct {
	ctx       context.Context
	dbNoTx    *gorm.DB
	machineId string
	tbPrefix  string
	objPrefix string
	key       string
	secret    string
	endpoint  string
	bucket    string
	expire    int64
}

func WithExportCtx(ctx context.Context) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).ctx = getExportCtx(ctx)
	}
}

func WithExportDbNoTx(db *gorm.DB) func(*ExportOptions) {
	return func(options *ExportOptions) {
		if db != nil {
			getExportOptionsOrSetDefault(options).dbNoTx = db
		}
	}
}

func WithExportMachineId(id string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).machineId = id
	}
}

func WithExportTbPrefix(prefix string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).tbPrefix = prefix
	}
}

func WithExportObjPrefix(prefix string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).objPrefix = prefix
	}
}

func WithExportKey(key string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).key = key
	}
}

func WithExportSecret(secret string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).secret = secret
	}
}

func WithExportEndpoint(endpoint string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		//if !strings.HasSuffix(endpoint, constant.DelayExportEndPointSuffix) {
		//	endpoint = endpoint + constant.DelayExportEndPointSuffix
		//}
		getExportOptionsOrSetDefault(options).endpoint = endpoint
	}
}

func WithExportBucket(bucket string) func(*ExportOptions) {
	return func(options *ExportOptions) {
		getExportOptionsOrSetDefault(options).bucket = bucket
	}
}

func WithExportExpire(min int64) func(*ExportOptions) {
	return func(options *ExportOptions) {
		if min > 0 {
			getExportOptionsOrSetDefault(options).expire = min
		}
	}
}

func getExportOptionsOrSetDefault(options *ExportOptions) *ExportOptions {
	if options == nil {
		return &ExportOptions{
			ctx:       getExportCtx(nil),
			tbPrefix:  constant.DelayExportTbPrefix,
			objPrefix: constant.DelayExportObjPrefix,
			machineId: fmt.Sprintf("%d", constant.One),
			endpoint:  "https://cq4oss.ctyunxs.cn",
			key:       "test",
			secret:    "test",
			bucket:    "test",
			expire:    constant.DelayExportObjExpire,
		}
	}
	return options
}

func getExportCtx(ctx context.Context) context.Context {
	if utils.InterfaceIsNil(ctx) {
		ctx = context.Background()
	}
	return context.WithValue(ctx, constant.LogSkipHelperCtxKey, false)
}

type QueueOptions struct {
	name            string
	redisUri        string
	redisPeriodKey  string
	retention       int
	maxRetry        int
	handler         func(ctx context.Context, t Task) error
	callback        string
	callbackTimeout int
	clearArchived   int
}

func WithQueueName(s string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).name = s
	}
}

func WithQueueRedisUri(s string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).redisUri = s
	}
}

func WithQueueRedisPeriodKey(s string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).redisPeriodKey = s
	}
}

func WithQueueRetention(second int) func(*QueueOptions) {
	return func(options *QueueOptions) {
		if second > 0 {
			getQueueOptionsOrSetDefault(options).retention = second
		}
	}
}

func WithQueueMaxRetry(count int) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).maxRetry = count
	}
}

func WithQueueHandler(fun func(ctx context.Context, t Task) error) func(*QueueOptions) {
	return func(options *QueueOptions) {
		if fun != nil {
			getQueueOptionsOrSetDefault(options).handler = fun
		}
	}
}

func WithQueueCallback(s string) func(*QueueOptions) {
	return func(options *QueueOptions) {
		getQueueOptionsOrSetDefault(options).callback = s
	}
}

func WithQueueCallbackTimeout(second int) func(*QueueOptions) {
	return func(options *QueueOptions) {
		if second > 0 {
			getQueueOptionsOrSetDefault(options).callbackTimeout = second
		}
	}
}

func WithQueueClearArchived(second int) func(*QueueOptions) {
	return func(options *QueueOptions) {
		if second > 0 {
			getQueueOptionsOrSetDefault(options).clearArchived = second
		}
	}
}

func getQueueOptionsOrSetDefault(options *QueueOptions) *QueueOptions {
	if options == nil {
		return &QueueOptions{
			name:            "delay",
			redisUri:        "redis://127.0.0.1:6379/0",
			redisPeriodKey:  "delay.queue.period",
			retention:       60,
			maxRetry:        3,
			callbackTimeout: 0,
			clearArchived:   300,
		}
	}
	return options
}

type QueueTaskOptions struct {
	uid       string
	name      string
	payload   string
	expr      string         // only period task
	in        *time.Duration // only once task
	at        *time.Time     // only once task
	now       bool           // only once task
	retention int            // only once task
	maxRetry  int
	timeout   int
}

func WithQueueTaskUuid(s string) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).uid = s
	}
}

func WithQueueTaskName(s string) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).name = s
	}
}

func WithQueueTaskPayload(s string) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).payload = s
	}
}

func WithQueueTaskExpr(s string) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).expr = s
	}
}

func WithQueueTaskIn(in time.Duration) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).in = &in
	}
}

func WithQueueTaskAt(at time.Time) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).at = &at
	}
}

func WithQueueTaskNow(flag bool) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).now = flag
	}
}

func WithQueueTaskRetention(second int) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		if second > 0 {
			getQueueTaskOptionsOrSetDefault(options).retention = second
		}
	}
}

func WithQueueTaskMaxRetry(count int) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		getQueueTaskOptionsOrSetDefault(options).maxRetry = count
	}
}

func WithQueueTaskTimeout(second int) func(*QueueTaskOptions) {
	return func(options *QueueTaskOptions) {
		if second > 0 {
			getQueueTaskOptionsOrSetDefault(options).timeout = second
		}
	}
}

func getQueueTaskOptionsOrSetDefault(options *QueueTaskOptions) *QueueTaskOptions {
	if options == nil {
		return &QueueTaskOptions{
			name:    "delay.queue.task",
			timeout: 0,
		}
	}
	return options
}
