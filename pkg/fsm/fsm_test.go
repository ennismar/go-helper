package fsm

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/req"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"testing"
)

var db *gorm.DB

func init() {
	dsn := "root:root@tcp(127.0.0.1:4306)/gin_web_stage?charset=utf8mb4&parseTime=True&loc=Local&timeout=10000ms"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "tb_",
		},
	})
	if err != nil {
		panic(fmt.Sprintf("[unit test]initialize mysql err: %v", err))
	}
	db = db.Debug()

	// db.AutoMigrate(
	// 	new(Test),
	// )
}

func TestMigrate(t *testing.T) {
	Migrate(WithDb(db), WithPrefix("tb_fsm"))
}

func TestFsm_CreateMachine(t *testing.T) {
	tx := db.Begin()
	f := New(WithDb(tx))
	f.CreateMachine(req.FsmCreateMachine{
		Name:                       "Leave Approval",
		SubmitterName:              "applicant",
		SubmitterEditFields:        "name,time,type",
		SubmitterConfirm:           req.NullUint(1),
		SubmitterConfirmEditFields: "status",
		Levels: []req.FsmCreateEvent{
			{
				Name:       "L1",
				Edit:       1,
				EditFields: "status",
				Users:      "4,5,6",
			},
			{
				Name:  "L2",
				Edit:  0,
				Roles: "4",
				Users: "8",
			},
			{
				Name:  "L3",
				Edit:  0,
				Roles: "5",
			},
		},
	})

	tx.Commit()
}

func TestFsm_SubmitLog(t *testing.T) {
	uid := "log1"
	tx := db.Begin()
	f := New(WithDb(tx))
	f.SubmitLog(req.FsmCreateLog{
		Category:        1,   // custom category
		Uuid:            uid, // unique str
		SubmitterUserId: 123, // submitter Id
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	tx.Commit()
}

func TestFsm_ApproveLog(t *testing.T) {
	uid := "log1"
	tx := db.Begin()
	f := New(WithDb(tx))
	// approved
	f.ApproveLog(req.FsmApproveLog{
		Category:       1,
		Uuid:           uid,
		ApprovalUserId: 4,
		Approved:       1,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// refused
	f.ApproveLog(req.FsmApproveLog{
		Category:        1,
		Uuid:            uid,
		ApprovalRoleId:  4,
		Approved:        2,
		ApprovalOpinion: "wrong info 1",
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// refused
	f.ApproveLog(req.FsmApproveLog{
		Category:        1,
		Uuid:            uid,
		ApprovalUserId:  4,
		Approved:        2,
		ApprovalOpinion: "wrong info 2",
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// resubmit
	f.ApproveLog(req.FsmApproveLog{
		Category:       1,
		Uuid:           uid,
		ApprovalUserId: 123,
		Approved:       1,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// approved
	f.ApproveLog(req.FsmApproveLog{
		Category:        1,
		Uuid:            uid,
		ApprovalUserId:  5,
		Approved:        1,
		ApprovalOpinion: "ok1",
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// approved
	f.ApproveLog(req.FsmApproveLog{
		Category:        1,
		Uuid:            uid,
		ApprovalRoleId:  4,
		Approved:        1,
		ApprovalOpinion: "ok2",
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// approved
	f.ApproveLog(req.FsmApproveLog{
		Category:        1,
		Uuid:            uid,
		ApprovalRoleId:  5,
		Approved:        1,
		ApprovalOpinion: "ok3",
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// submitter confirmed
	f.ApproveLog(req.FsmApproveLog{
		Category:       1,
		Uuid:           uid,
		ApprovalUserId: 123,
		Approved:       1,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	tx.Commit()
}

func TestFsm_ApproveLog1(t *testing.T) {
	uid := "log2"
	tx := db.Begin()
	f := New(WithDb(tx))
	f.SubmitLog(req.FsmCreateLog{
		Category:        1,
		Uuid:            uid,
		SubmitterUserId: 234,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	// other people cancel
	f.ApproveLog(req.FsmApproveLog{
		Category:       1,
		Uuid:           uid,
		ApprovalUserId: 5,
		Approved:       3,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	// submitter cancel
	f.ApproveLog(req.FsmApproveLog{
		Category:       1,
		Uuid:           uid,
		ApprovalUserId: 234,
		Approved:       3,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	f.SubmitLog(req.FsmCreateLog{
		Category:        1,
		Uuid:            uid,
		SubmitterRoleId: 567,
		SubmitterUserId: 456,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	// submitter role cancel
	f.ApproveLog(req.FsmApproveLog{
		Category:       1,
		Uuid:           uid,
		ApprovalRoleId: 567,
		Approved:       3,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	tx.Commit()
}

func TestFsm_CancelLogs(t *testing.T) {
	tx := db.Begin()
	f := New(WithDb(tx))
	f.SubmitLog(req.FsmCreateLog{
		Category:        1,
		Uuid:            "log3",
		SubmitterUserId: 123,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	f.SubmitLog(req.FsmCreateLog{
		Category:        1,
		Uuid:            "log4",
		SubmitterUserId: 234,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}
	f.SubmitLog(req.FsmCreateLog{
		Category:        1,
		Uuid:            "log5",
		SubmitterUserId: 345,
	})
	if f.Error != nil {
		fmt.Println(f.Error)
	}

	f.CancelLog(1)
	tx.Commit()
}

func TestFsm_FindPendingLogsByApprover(t *testing.T) {
	tx := db.Begin()
	f := New(WithDb(tx))
	fmt.Println(f.FindPendingLogByApprover(&req.FsmPendingLog{
		ApprovalRoleId: 1,
		ApprovalUserId: 2,
		Category:       1,
	}))
	tx.Commit()
}

func TestFsm_FindLogs(t *testing.T) {
	tx := db.Begin()
	f := New(WithDb(tx))
	fmt.Println(f.FindLog(req.FsmLog{
		Category: 1,
		Uuid:     "log1",
	}))
	tx.Commit()
}

func TestFsm_GetLogTrack(t *testing.T) {
	tx := db.Begin()
	f := New(WithDb(tx))
	list := f.FindLog(req.FsmLog{
		Category: 1,
		Uuid:     "log2",
	})
	fmt.Println(f.FindLogTrack(list))
	tx.Commit()
}
