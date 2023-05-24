package router

import v1 "github.com/ennismar/go-helper/api/v1"

func (rt Router) OperationLog() {
	router1 := rt.Casbin("/operation/log")
	router1.GET("/list", v1.FindOperationLog(rt.ops.v1Ops...))
	router1.DELETE("/delete/batch", v1.BatchDeleteOperationLogByIds(rt.ops.v1Ops...))
}
