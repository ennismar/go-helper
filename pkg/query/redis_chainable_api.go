package query

import (
	"regexp"
	"strings"
)

var tableRegexp = regexp.MustCompile(`(?i).+? AS (\w+)\s*(?:$|,)`)

// FromString set json str
func (rd *Redis) FromString(str string) *Redis {
	ins := rd.getInstance()
	ins.Statement.FromString(str)
	ins.Statement.Dest = nil
	return ins
}

// Table set table name
func (rd *Redis) Table(name string, args ...interface{}) (ins *Redis) {
	ins = rd.getInstance()
	if strings.Contains(name, " ") || strings.Contains(name, "`") || len(args) > 0 {
		if results := tableRegexp.FindStringSubmatch(name); len(results) == 2 {
			ins.Statement.Table = rd.ops.namingStrategy.TableName(results[1])
			return
		}
	}

	ins.Statement.Table = rd.ops.namingStrategy.TableName(name)
	return
}

// Preload preload column
func (rd *Redis) Preload(column string) *Redis {
	return rd.getInstance().Statement.Preload(column).DB
}

// Where where condition
func (rd *Redis) Where(key, cond string, val interface{}) *Redis {
	return rd.getInstance().Statement.Where(key, cond, val).DB
}

// Order sort condition
func (rd *Redis) Order(key string) *Redis {
	return rd.getInstance().Statement.Order(key).DB
}

// Limit limit condition
func (rd *Redis) Limit(limit int) *Redis {
	ins := rd.getInstance()
	ins.Statement.limit = limit
	return ins
}

// Offset offset condition
func (rd *Redis) Offset(offset int) *Redis {
	ins := rd.getInstance()
	ins.Statement.offset = offset
	return ins
}
