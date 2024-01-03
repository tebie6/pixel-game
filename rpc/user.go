package rpc

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/tools/log"
	"strconv"
)

// 游戏游客用户登录
func gameVisitorLogin(c *gin.Context) {

	// 可在此处增加用户指纹逻辑，防止用户Token丢失
	// 注册用户
	registerData := models.GameUser{}
	userId, err := models.CreateGameUser(&registerData)
	if err != nil {
		log.Error("user", fmt.Sprintf("CreateGameUser error: %v", err))
		apiError(c, ApiErrMsg, "系统错误")
		return
	}

	userInfo, err := models.GetGameUserById(userId)
	if err != nil {
		log.Error("user", fmt.Sprintf("GetUserByUsername error: %v", err))
		apiError(c, ApiErrMsg, "系统错误")
		return
	}
	if userInfo == nil {
		apiError(c, ApiErrMsg, "系统错误")
		return
	}

	// 创建token
	token := jwt.New(jwt.SigningMethodHS256)
	// 创建一个token的声明
	claims := token.Claims.(jwt.MapClaims)
	// 可以添加自定义的用户ID或其他信息
	claims["user_id"] = strconv.FormatInt(userInfo.Id, 10)
	claims["created_at"] = userInfo.CreatedAt

	t, err := token.SignedString([]byte("tebie6.com"))
	if err != nil {
		log.Error("user", fmt.Sprintf("create token error: %v", err))
		apiError(c, ApiErrMsg, "登录失败 1001")
		return
	}

	userInfo.AccessToken = t
	err = models.UpdateGameUser(userInfo)
	if err != nil {
		log.Error("user", fmt.Sprintf("SaveUserLoginTime error: %v", err))
		apiError(c, ApiErrMsg, "登录失败")
		return
	}

	var resp = map[string]interface{}{
		"access_token": t,
	}
	apiSuccess(c, resp)

	return
}
