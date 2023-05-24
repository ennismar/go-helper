package req

import "github.com/ennismar/go-helper/pkg/resp"

type Api struct {
	Method   string `json:"method" form:"method"`
	Path     string `json:"path" form:"path"`
	Category string `json:"category" form:"category"`
	resp.Page
}

type CreateApi struct {
	Method       string   `json:"method" validate:"required"`
	Path         string   `json:"path" validate:"required"`
	Category     string   `json:"category" validate:"required"`
	Desc         string   `json:"desc"`
	Title        string   `json:"title"`
	RoleIds      []uint   `json:"roleIds"`
	RoleKeywords []string `json:"roleKeywords"`
}

func (s CreateApi) FieldTrans() map[string]string {
	m := make(map[string]string, 0)
	m["Method"] = "method"
	m["Path"] = "path"
	m["Category"] = "category"
	return m
}

type UpdateApi struct {
	Method   *string `json:"method"`
	Path     *string `json:"path"`
	Category *string `json:"category"`
	Desc     *string `json:"desc"`
	Title    *string `json:"title"`
	RoleIds  []uint  `json:"roleIds"`
}
