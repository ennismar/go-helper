package delay

import (
	"fmt"
	"github.com/ennismar/go-helper/pkg/req"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"testing"
	"time"
)

func TestNewExport(t *testing.T) {
	db, _ := gorm.Open(mysql.Open("root:rO0tSDfjkuisdfDFuio@tcp(39.106.224.169:3306)/gsgc_web_prod?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=10000ms"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "tb_",
			SingularTable: true,
		},
	})
	MigrateExport(
		WithExportDbNoTx(db),
	)
	ex := NewExport(
		WithExportDbNoTx(db),
	)
	ex.Start("uuid3", "export 1", "category 1", "start")
	i := 0
	for {
		if i >= 100 {
			break
		}
		ex.Pending("uuid3", fmt.Sprintf("%d%%", i))
		i++
		time.Sleep(10 * time.Millisecond)
	}
	ex.End("uuid3", fmt.Sprintf("%d%%", i), "/tmp/test.txt")

	fmt.Println(ex.FindHistory(&req.DelayExportHistory{}))
}
