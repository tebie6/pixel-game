package rpc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tebie6/pixel-game/conf"
	"github.com/tebie6/pixel-game/tools/log"
	"net/http"
)

const (
	ApiSuccess = 0
	ApiErrMsg  = 500
)

type ApiRet struct {
	Code int32                  `json:"code"`
	Data map[string]interface{} `json:"data"`
	Msg  string                 `json:"msg"`
}

func InitRpc() {

	r := gin.New()

	r.Static("/assets", conf.GetConfigString("app", "static_path"))

	r = routerRegister(r)

	rpcListenAddress := fmt.Sprintf("%s:%s",
		conf.GetConfigString("rpc", "host"),
		conf.GetConfigString("rpc", "port"),
	)
	err := r.Run(rpcListenAddress)
	if err != nil {
		log.Error("", "start server failed %s", err.Error())
	}
}

func routerRegister(r *gin.Engine) *gin.Engine {

	r.GET("/api/", test)

	// user
	r.POST("/api/visitor/login", gameVisitorLogin)

	// pixel
	r.Any("/api/pixel/list", pixelList)
	r.Any("/api/pixel/repairContent", repairContent)
	r.POST("/api/pixel/online/list", pixelOnlineUserList)
	r.POST("/api/pixel/chat/list", pixelChatList)
	r.GET("/api/pixel/addSensitiveWord", addSensitiveWord)
	r.POST("/api/pixel/errorReporting", errorReporting)

	return r
}

func apiSuccess(c *gin.Context, data map[string]interface{}) {
	ret := &ApiRet{
		Code: ApiSuccess,
		Data: data,
		Msg:  "success",
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(http.StatusOK, ret)
	return
}

func apiError(c *gin.Context, code int32, msg string) {
	ret := &ApiRet{
		Code: code,
		Msg:  msg,
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(http.StatusOK, ret)
	return
}
