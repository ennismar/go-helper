package interceptor

import (
	"context"
	"github.com/ennismar/go-helper/pkg/log"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func Exception(options ...func(*ExceptionOptions)) grpc.UnaryServerInterceptor {
	ops := getExceptionOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	return grpc_recovery.UnaryServerInterceptor(
		grpc_recovery.WithRecoveryHandlerContext(
			func(ctx context.Context, p interface{}) (err error) {
				err = errors.Errorf("%v", p)
				log.WithContext(ctx).WithError(err).Error("runtime exception")
				return
			},
		),
	)
}
