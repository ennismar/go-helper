package oss

import (
	"context"
	"github.com/ennismar/go-helper/pkg/utils"
)

type MinioOptions struct {
	ctx      context.Context
	endpoint string
	accessId string
	secret   string
	https    bool
}

func WithMinioCtx(ctx context.Context) func(*MinioOptions) {
	return func(options *MinioOptions) {
		if !utils.InterfaceIsNil(ctx) {
			getMinioOptionsOrSetDefault(options).ctx = ctx
		}
	}
}

func WithMinioEndpoint(endpoint string) func(*MinioOptions) {
	return func(options *MinioOptions) {
		getMinioOptionsOrSetDefault(options).endpoint = endpoint
	}
}

func WithMinioAccessId(accessId string) func(*MinioOptions) {
	return func(options *MinioOptions) {
		getMinioOptionsOrSetDefault(options).accessId = accessId
	}
}

func WithMinioSecret(secret string) func(*MinioOptions) {
	return func(options *MinioOptions) {
		getMinioOptionsOrSetDefault(options).secret = secret
	}
}

func WithMinioHttps(flag bool) func(*MinioOptions) {
	return func(options *MinioOptions) {
		getMinioOptionsOrSetDefault(options).https = flag
	}
}

func getMinioOptionsOrSetDefault(options *MinioOptions) *MinioOptions {
	if options == nil {
		return &MinioOptions{
			ctx: context.Background(),
		}
	}
	return options
}
