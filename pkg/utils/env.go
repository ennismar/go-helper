package utils

import (
	"fmt"
	"github.com/piupuer/go-helper/pkg/log"
	"os"
	"strconv"
	"strings"
)

// EnvToInterface parse environment variables to interface
func EnvToInterface(options ...func(*EnvOptions)) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	m := make(map[string]interface{}, 0)
	Struct2StructByJson(ops.obj, &m)
	newMap := envToInterface(m, ops.prefix, ops.format)
	Struct2StructByJson(newMap, ops.obj)
}

func envToInterface(m map[string]interface{}, prefix string, fun func(key string, val interface{}) string) map[string]interface{} {
	newMap := make(map[string]interface{}, 0)
	// json types: string/bool/float64
	for key, item := range m {
		newKey := strings.ToUpper(SnakeCase(key))
		if prefix != "" {
			newKey = strings.ToUpper(fmt.Sprintf("%s_%s", SnakeCase(prefix), SnakeCase(key)))
		}
		switch item.(type) {
		case map[string]interface{}:
			itemM, _ := item.(map[string]interface{})
			newMap[key] = envToInterface(itemM, newKey, fun)
			continue
		case string:
			env := strings.TrimSpace(os.Getenv(newKey))
			if env != "" {
				newMap[key] = env
				log.Info("[env to interface]get %v", fun(newKey, newMap[key]))
				continue
			}
		case bool:
			env := strings.TrimSpace(os.Getenv(newKey))
			if env != "" {
				itemB, ok := item.(bool)
				b, err := strconv.ParseBool(env)
				if ok && err == nil {
					if itemB && !b {
						newMap[key] = false
						log.Info("[env to interface]get %v", fun(newKey, newMap[key]))
						continue
					} else if !itemB && b {
						newMap[key] = true
						log.Info("[env to interface]get %v", fun(newKey, newMap[key]))
						continue
					}
				}
			}
		case float64:
			e := strings.TrimSpace(os.Getenv(newKey))
			if e != "" {
				v, err := strconv.ParseFloat(e, 64)
				if err == nil {
					newMap[key] = v
					log.Info("[env to interface]get %v", fun(newKey, newMap[key]))
					continue
				}
			}
		}
		// no difference
		newMap[key] = item
	}
	return newMap
}
