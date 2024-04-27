package utils

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetQueryMap(c *gin.Context, paramsKey []string) (map[string]string, error) {
	paramsMap := make(map[string]string)
	for _, key := range paramsKey {
		if value, ok := c.GetQuery(key); !ok {
			return nil, errors.New(fmt.Sprintln("key: ", key, " not found"))
		} else {
			paramsMap[key] = value
		}
	}
	return paramsMap, nil
}
