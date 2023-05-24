package query

import (
	"github.com/piupuer/go-helper/pkg/log"
	"github.com/piupuer/go-helper/pkg/resp"
	"reflect"
)

// Find like gorm.Find
func (rd Redis) Find(dest interface{}) *Redis {
	ins := rd.getInstance()
	if !ins.check() {
		return ins
	}
	ins.Statement.Dest = dest
	ins.Statement.Model = dest
	ins.Statement.count = false
	ins.beforeQuery(ins).findByTableName(ins.Statement.Table)
	return ins
}

// First like gorm.First
func (rd Redis) First(dest interface{}) *Redis {
	ins := rd.getInstance()
	ins.Statement.limit = 1
	ins.Statement.first = true
	ins.Find(dest)
	return ins
}

// Count like gorm.Count
func (rd Redis) Count(count *int64) *Redis {
	ins := rd.getInstance()
	ins.Statement.Dest = count
	ins.Statement.count = true
	if !ins.check() {
		*count = 0
	}
	ins.Statement.Dest = count
	*count = int64(ins.beforeQuery(ins).findByTableName(ins.Statement.Table).Count())
	return ins
}

func (rd Redis) FindWithPage(q *Redis, page *resp.Page, model interface{}) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr || (rv.IsNil() || rv.Elem().Kind() != reflect.Slice) {
		log.WithContext(rd.Ctx).Warn("model must be a pointer")
		return
	}

	if !page.NoPagination {
		q.Count(&page.Total)
		if page.Total > 0 {
			limit, offset := page.GetLimit()
			q.Limit(limit).Offset(offset).Find(model)
		}
	} else {
		q.Find(model)
		page.Total = int64(rv.Elem().Len())
		page.GetLimit()
	}
}
