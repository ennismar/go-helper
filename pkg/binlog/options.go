package binlog

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/piupuer/go-helper/pkg/utils"
	"gorm.io/gorm"
)

type Options struct {
	ctx       context.Context
	dsn       *mysql.Config
	db        *gorm.DB
	redis     redis.UniversalClient
	models    []interface{}
	serverId  uint32
	binlogPos string
}

func WithCtx(ctx context.Context) func(*Options) {
	return func(options *Options) {
		if !utils.InterfaceIsNil(ctx) {
			getOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithDsn(dsn *mysql.Config) func(*Options) {
	return func(options *Options) {
		if dsn != nil {
			getOptionsOrSetDefault(options).dsn = dsn
		}
	}
}

func WithDb(db *gorm.DB) func(*Options) {
	return func(options *Options) {
		if db != nil {
			getOptionsOrSetDefault(options).db = db
		}
	}
}

func WithRedis(rd redis.UniversalClient) func(*Options) {
	return func(options *Options) {
		if rd != nil {
			getOptionsOrSetDefault(options).redis = rd
		}
	}
}

func WithModels(models ...interface{}) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).models = append(getOptionsOrSetDefault(options).models, models...)
	}
}

func WithServerId(serverId uint32) func(*Options) {
	return func(options *Options) {
		if serverId > 0 {
			getOptionsOrSetDefault(options).serverId = serverId
		}
	}
}

func WithBinlogPos(key string) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).binlogPos = key
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			ctx:       context.Background(),
			serverId:  100,
			binlogPos: "mysql_binlog_pos",
		}
	}
	return options
}
