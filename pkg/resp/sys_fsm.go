package resp

type FsmApprovalLog struct {
	Uuid     string `json:"uuid"`
	Category uint   `json:"category"`
	End      uint   `json:"end"`      // is ended?
	Confirm  uint   `json:"confirm"`  // is waiting submitter confirm?
	Resubmit uint   `json:"resubmit"` // is waiting submitter resubmit?
	Cancel   uint   `json:"cancel"`   // is submitter canceled?
}

type FsmApprovingLog struct {
	Base
	Uuid             string `json:"uuid"`
	Category         uint   `json:"category"`
	SubmitterRoleId  uint   `json:"submitterRoleId"`
	SubmitterRole    Role   `json:"submitterRole"`
	SubmitterUserId  uint   `json:"submitterUserId"`
	SubmitterUser    User   `json:"submitterUser"`
	PrevDetail       string `json:"prevDetail"`
	Detail           string `json:"detail"`
	Remark           string `json:"remark"`
	Confirm          uint   `json:"confirm"`
	Resubmit         uint   `json:"resubmit"`
	CanApprovalRoles []Role `json:"canApprovalRoles"`
	CanApprovalUsers []User `json:"canApprovalUsers"`
}

type FsmLogTrack struct {
	Time
	Name       string `json:"name"`
	Opinion    string `json:"opinion"`
	Status     uint   `json:"status"`
	End        uint   `json:"end"`
	Cancel     uint   `json:"cancel"`
	Resubmit   uint   `json:"resubmit"`
	Confirm    uint   `json:"confirm"`
	Permission uint   `json:"permission"`
}

type FsmLogSubmitterDetail struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Val  string `json:"val"`
}

type FsmMachine struct {
	Base
	Category                   uint   `json:"category"`
	Name                       string `json:"name"`
	SubmitterName              string `json:"submitterName"`
	SubmitterEditFields        string `json:"submitterEditFields"`
	SubmitterConfirm           uint   `json:"submitterConfirm"`
	SubmitterConfirmEditFields string `json:"submitterConfirmEditFields"`
	EventsJson                 string `json:"eventsJson"`
}
