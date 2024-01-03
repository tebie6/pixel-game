package rpc

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func test(c *gin.Context) {

	fmt.Println(c.Request)

	apiSuccess(c, map[string]interface{}{
		"status" : "ok",
	})
}