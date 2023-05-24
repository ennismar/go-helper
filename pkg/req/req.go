package req

import (
	"context"
	"github.com/ennismar/go-helper/pkg/resp"
	"github.com/ennismar/go-helper/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

// ShouldBind bind request param
func ShouldBind(c *gin.Context, r interface{}) {
	err := c.ShouldBind(r)
	if err != nil {
		resp.FailWithMsg("%s: %v", resp.InvalidParameterMsg, err)
	}
}

// UintId get uint path id
func UintId(c *gin.Context) uint {
	i := c.Param("id")
	id := utils.Str2Uint(i)
	if id == 0 {
		resp.CheckErr("invalid path id: %s", i)
	}
	return id
}

// UintIdWithErr get uint path id with err
func UintIdWithErr(c *gin.Context) (id uint, err error) {
	i := c.Param("id")
	id = utils.Str2Uint(i)
	if id == 0 {
		err = errors.Errorf("invalid path id")
		return
	}
	return
}

// UintIds get uint path ids
func UintIds(c *gin.Context) (arr []uint) {
	i := c.Param("ids")
	arr = utils.Str2UintArr(i)
	if len(arr) == 0 {
		resp.CheckErr("invalid path ids: %s", i)
	}
	return
}

// UintIdsWithErr get uint path ids
func UintIdsWithErr(c *gin.Context) (arr []uint, err error) {
	i := c.Param("ids")
	arr = utils.Str2UintArr(i)
	if len(arr) == 0 {
		err = errors.Errorf("invalid path ids: %s", i)
		return
	}
	return
}

// Validate validate request param
func Validate(ctx context.Context, r interface{}, trans map[string]string, options ...func(*ValidateOptions)) {
	ops := getValidateOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	err := validate(ops.validator.Struct(r), trans, *ops)
	if err != nil {
		resp.FailWithMsg("%s: %v", resp.IllegalParameterMsg, err)
	}
}

// ValidateWithErr validate request param return err
func ValidateWithErr(ctx context.Context, r interface{}, trans map[string]string, options ...func(*ValidateOptions)) error {
	ops := getValidateOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	err := validate(ops.validator.Struct(r), trans, *ops)
	return err
}

func validate(err error, custom map[string]string, ops ValidateOptions) (e error) {
	if err == nil {
		return
	}
	errs := err.(validator.ValidationErrors)
	for _, item := range errs {
		tranStr := item.Translate(ops.translator)
		names := strings.Split(item.Namespace(), ".")
		// deep names
		if len(names) > 1 {
			if v, ok := custom[strings.Join(names[1:], ".")]; ok {
				return errors.Errorf(strings.Replace(tranStr, item.Field(), v, 1))
			}
		}
		// check whether it is in custom
		if v, ok := custom[item.Field()]; ok {
			return errors.Errorf(strings.Replace(tranStr, item.Field(), v, 1))
		} else {
			return errors.Errorf(tranStr)
		}
	}
	return
}
