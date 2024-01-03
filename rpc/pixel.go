package rpc

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tebie6/pixel-game/conf"
	"github.com/tebie6/pixel-game/services"
	"github.com/tebie6/pixel-game/tools/log"
	"github.com/tebie6/pixel-game/tools/rpchelper"
	"github.com/tebie6/pixel-game/tools/sensitive"
)

// pixelList 画布列表
func pixelList(c *gin.Context) {
	pixelService := services.PixelService{}
	res, err := pixelService.GetList()
	if err != nil {
		log.Error("pixel", fmt.Sprintf("pixelList GetList err: %s", err))
		apiError(c, ApiErrMsg, "获取画布列表失败")
		return
	}

	apiSuccess(c, map[string]interface{}{
		"list": res,
	})
}

// repairContent 恢复画布
func repairContent(c *gin.Context) {
	password := rpchelper.RequestParameterString(c, "password")
	if password != conf.GetConfigString("app", "password") {
		apiSuccess(c, map[string]interface{}{
			"status": 0,
		})
		return
	}

	pixelService := services.PixelService{}
	_ = pixelService.RepairContent()

	apiSuccess(c, map[string]interface{}{
		"status": 1,
	})
}

// pixelOnlineUserList 在线用户列表
func pixelOnlineUserList(c *gin.Context) {
	userService := services.UserService{}
	res := userService.GetOnlineUserList()

	apiSuccess(c, map[string]interface{}{
		"list": res,
	})
}

// pixelChatList 聊天列表
func pixelChatList(c *gin.Context) {
	pixelService := services.PixelService{}
	res, _ := pixelService.GetChatList()

	apiSuccess(c, map[string]interface{}{
		"list": res,
	})
}

// addSensitiveWord 添加敏感词
func addSensitiveWord(c *gin.Context) {
	word := rpchelper.RequestParameterString(c, "word")
	if len(word) == 0 {
		apiSuccess(c, map[string]interface{}{
			"status": 0,
		})
		return
	}
	password := rpchelper.RequestParameterString(c, "password")
	if password != conf.GetConfigString("app", "password") {
		apiSuccess(c, map[string]interface{}{
			"status": 0,
		})
		return
	}

	// 动态添加敏感词
	sensitive.SensitiveFilter.AddWord(word)

	apiSuccess(c, map[string]interface{}{
		"word":   word,
		"status": 1,
	})
}

// errorReporting 错误上报
func errorReporting(c *gin.Context) {
	message := rpchelper.RequestParameterString(c, "message")
	source := rpchelper.RequestParameterString(c, "source")
	lineno := rpchelper.RequestParameterString(c, "lineno")
	colno := rpchelper.RequestParameterString(c, "colno")
	stack := rpchelper.RequestParameterString(c, "stack")
	accessToken := rpchelper.RequestParameterString(c, "access_token")
	if len(accessToken) == 0 {
		apiError(c, ApiErrMsg, "access_token is empty")
		return
	}

	pixelService := services.PixelService{}
	err := pixelService.ErrorReporting(message, source, lineno, colno, stack, accessToken)
	if err != nil {
		log.Error("pixel", fmt.Sprintf("errorReporting err:%s", err))
		apiError(c, ApiErrMsg, "请求错误")
		return
	}

	apiSuccess(c, map[string]interface{}{})
}
