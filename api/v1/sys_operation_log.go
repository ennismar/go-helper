package v1

import (
	"github.com/ennismar/go-helper/ms"
	"github.com/ennismar/go-helper/pkg/query"
	"github.com/ennismar/go-helper/pkg/req"
	"github.com/ennismar/go-helper/pkg/resp"
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/gin-gonic/gin"
)

// FindOperationLog
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *OperationLog
// @Description FindOperationLog
// @Param params query req.OperationLog true "params"
// @Router /operation/log/list [GET]
func FindOperationLog(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "FindOperationLog"))
		defer span.End()
		var r req.OperationLog
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		list := q.FindOperationLog(&r)
		resp.SuccessWithPageData(list, &[]resp.OperationLog{}, r.Page)
	}
}

// BatchDeleteOperationLogByIds
// @Security Bearer
// @Accept json
// @Produce json
// @Success 201 {object} resp.Resp "success"
// @Tags *OperationLog
// @Description BatchDeleteOperationLogByIds
// @Param ids body req.Ids true "ids"
// @Router /operation/log/delete/batch [DELETE]
func BatchDeleteOperationLogByIds(options ...func(*Options)) gin.HandlerFunc {
	ops := ParseOptions(options...)
	return func(c *gin.Context) {
		ctx := tracing.RealCtx(c)
		_, span := tracer.Start(ctx, tracing.Name(tracing.Rest, "BatchDeleteOperationLogByIds"))
		defer span.End()
		if !ops.operationAllowedToDelete {
			resp.CheckErr("this feature has been turned off by the administrator")
		}
		var r req.Ids
		req.ShouldBind(c, &r)
		ops.addCtx(c)
		q := query.NewMySql(ops.dbOps...)
		err := q.DeleteByIds(r.Uints(), new(ms.SysOperationLog))
		resp.CheckErr(err)
		resp.Success()
	}
}
