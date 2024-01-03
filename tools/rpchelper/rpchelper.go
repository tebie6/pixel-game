package rpchelper

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

// 获取字符串值
func RequestParameterString(c *gin.Context, key string) string {
	v := c.Request.FormValue(key)
	return v
}

// 获取整数值
func RequestParameterInt(c *gin.Context, key string) (int64, bool) {
	v := RequestParameterString(c, key)

	if v == "" {
		return 0, false
	}

	i, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		return 0, false
	}

	return i, true
}

// 获取浮点值
func RequestParameterFloat(c *gin.Context, key string) (float64, bool) {
	v := RequestParameterString(c, key)

	if v == "" {
		return 0, false
	}

	i, e := strconv.ParseFloat(v, 64)
	if e != nil {
		return 0, false
	}

	return i, true
}
