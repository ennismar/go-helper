package query

import (
	"github.com/casbin/casbin/v2"
	"github.com/ennismar/go-helper/ms"
	"github.com/ennismar/go-helper/pkg/log"
	"github.com/ennismar/go-helper/pkg/tracing"
	"github.com/ennismar/go-helper/pkg/utils"
	"github.com/pkg/errors"
)

func (my MySql) FindRoleCasbin(c ms.SysRoleCasbin) []ms.SysRoleCasbin {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "FindRoleCasbin"))
	defer span.End()
	cs := make([]ms.SysRoleCasbin, 0)
	if my.ops.enforcer == nil {
		log.WithContext(my.Ctx).Warn("casbin enforcer is empty")
		return cs
	}
	policies := my.ops.enforcer.GetFilteredPolicy(0, c.Keyword, c.Path, c.Method)
	for _, policy := range policies {
		cs = append(cs, ms.SysRoleCasbin{
			Keyword: policy[0],
			Path:    policy[1],
			Method:  policy[2],
		})
	}
	return cs
}

func (my MySql) CreateRoleCasbin(c ms.SysRoleCasbin) (bool, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "CreateRoleCasbin"))
	defer span.End()
	if my.ops.enforcer == nil {
		return false, errors.Errorf("casbin enforcer is empty")
	}
	return my.ops.enforcer.AddPolicy(c.Keyword, c.Path, c.Method)
}

func (my MySql) BatchCreateRoleCasbin(cs []ms.SysRoleCasbin) (bool, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "BatchCreateRoleCasbin"))
	defer span.End()
	rules := make([][]string, 0)
	if my.ops.enforcer == nil {
		return false, errors.Errorf("casbin enforcer is empty")
	}
	for _, c := range cs {
		rules = append(rules, []string{
			c.Keyword,
			c.Path,
			c.Method,
		})
	}
	return my.ops.enforcer.AddPolicies(rules)
}

func (my MySql) DeleteRoleCasbin(c ms.SysRoleCasbin) (bool, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "DeleteRoleCasbin"))
	defer span.End()
	if my.ops.enforcer == nil {
		return false, errors.Errorf("casbin enforcer is empty")
	}
	return my.ops.enforcer.RemovePolicy(c.Keyword, c.Path, c.Method)
}

func (my MySql) BatchDeleteRoleCasbin(cs []ms.SysRoleCasbin) (bool, error) {
	_, span := tracer.Start(my.Ctx, tracing.Name(tracing.Db, "BatchDeleteRoleCasbin"))
	defer span.End()
	if my.ops.enforcer == nil {
		return false, errors.Errorf("casbin enforcer is empty")
	}
	rules := make([][]string, 0)
	for _, c := range cs {
		rules = append(rules, []string{
			c.Keyword,
			c.Path,
			c.Method,
		})
	}
	return my.ops.enforcer.RemovePolicies(rules)
}

func FindCasbinByRoleKeyword(enforcer *casbin.Enforcer, roleKeyword string) (rp []ms.SysCasbin) {
	rp = make([]ms.SysCasbin, 0)
	list := make([][]string, 0)
	if roleKeyword != "" {
		// filter rules by keyword
		list = enforcer.GetFilteredPolicy(0, roleKeyword)
	} else {
		list = enforcer.GetFilteredPolicy(0)
	}

	var added []string
	for _, v := range list {
		if !utils.Contains(added, v[1]+v[2]) {
			rp = append(rp, ms.SysCasbin{
				PType: "p",
				V1:    v[1],
				V2:    v[2],
			})
			added = append(added, v[1]+v[2])
		}
	}
	return
}
