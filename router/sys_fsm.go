package router

import v1 "github.com/piupuer/go-helper/api/v1"

func (rt Router) Fsm() {
	router1 := rt.Casbin("/fsm")
	router2 := rt.CasbinAndIdempotence("/fsm")
	router1.GET("/list", v1.FindFsm(rt.ops.v1Ops...))
	router2.POST("/create", v1.CreateFsm(rt.ops.v1Ops...))
	router1.PATCH("/update/:id", v1.UpdateFsmById(rt.ops.v1Ops...))
	router1.DELETE("/delete/batch", v1.BatchDeleteFsmByIds(rt.ops.v1Ops...))
	router1.GET("/log/approving/list", v1.FindFsmApprovingLog(rt.ops.v1Ops...))
	router1.GET("/log/track", v1.FindFsmLogTrack(rt.ops.v1Ops...))
	router1.GET("/log/submitter/detail", v1.GetFsmLogSubmitterDetail(rt.ops.v1Ops...))
	router1.PATCH("/log/submitter/detail", v1.UpdateFsmLogSubmitterDetail(rt.ops.v1Ops...))
	router1.PATCH("/log/approve", v1.FsmApproveLog(rt.ops.v1Ops...))
	router1.PATCH("/log/cancel", v1.FsmCancelLogByUuids(rt.ops.v1Ops...))
}
