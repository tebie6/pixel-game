package socketctrls

import (
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/netws"
	"github.com/tebie6/pixel-game/services"
	"github.com/tebie6/pixel-game/tools/log"
	"github.com/tebie6/pixel-game/tools/sensitive"
	"golang.org/x/time/rate"
	"runtime"
	"strconv"
	"time"
)

type PixelGameControlle struct {
}

func (inst *PixelGameControlle) Init() map[string]func(*models.Subscriber, interface{}) {
	return map[string]func(*models.Subscriber, interface{}){
		netws.WSActionAuthReq:      inst.AuthAction,
		netws.WSActionPixelAddReq:  inst.AddPixAction,
		netws.WSActionPixelChatReq: inst.PixelChat,
	}
}

// AuthAction 验证权限
func (inst *PixelGameControlle) AuthAction(sub *models.Subscriber, req interface{}) {
	reqData := req.(*netws.AuthReq)

	if len(reqData.Token) == 0 {
		pushAuthMsg(sub, "bye bye", "", "")
		netws.Leave(sub)
		return
	}

	// 验证token
	userInfo, err := models.GetUserByAccessToken(reqData.Token)
	if err != nil {
		pushAuthMsg(sub, "10001", "", "")
		netws.Leave(sub)
		return
	}

	if userInfo == nil {
		pushAuthMsg(sub, "token error", "", "")
		netws.Leave(sub)
		return
	}

	// 禁用用户
	if userInfo.Status == 0 {
		pushAuthMsg(sub, "blocked", "", "")
		netws.Leave(sub)
		return
	}

	// 保存身份连接信息
	sub.Uid = userInfo.Id

	// 记录该用户总连接量
	netws.UserWsConnectionCounter.AddConnection(sub.Uid)
	if netws.UserWsConnectionCounter.GetConnectionCount(sub.Uid) > 3 {
		pushAuthMsg(sub, "too many connect", "", "")
		netws.Leave(sub)
		return
	}

	// 创建一个新的令牌桶，每秒放入 5 个令牌，桶的容量为 50 个令牌
	sub.Limiter = rate.NewLimiter(5, 50)

	pushAuthMsg(sub, "", userInfo.Nickname, sub.Uuid)

	// 广播数据 - 当前人数
	pushBroadcast(map[string]string{
		"action":   "online",
		"online":   strconv.FormatInt(netws.ConnectAmount, 10),
		"id":       sub.Uuid,
		"type":     "join",
		"nickname": userInfo.Nickname,
	})

	// 加载方式 0:ws推送内容、1:图片加载
	if reqData.LoadMode == 0 {
		// 渲染列表
		pixelService := services.PixelService{}
		res, _ := pixelService.GetList()

		total := len(res) // 总数量
		chunkSize := 128
		if reqData.Size != 0 {
			chunkSize = reqData.Size
		}
		pushTotal := 0 // 推送数量

		chunkMap := make(map[string]int16)
		for key, color := range res {
			chunkMap[key] = int16(color)
			// 当达到chunkSize或处理完所有数据时，发送数据
			if len(chunkMap) == chunkSize {
				pushTotal += len(chunkMap)
				pushPixelMsgUpdateCanvas(sub, chunkMap, pushTotal, total)
				chunkMap = make(map[string]int16)
			}
		}

		// 处理剩余的数据
		if len(chunkMap) > 0 {
			pushTotal += len(chunkMap)
			pushPixelMsgUpdateCanvas(sub, chunkMap, pushTotal, total)
		}

		// 通知前端加载完成
		pushPixelMsg(sub, netws.PixelAddReq{
			Action: "pixel_add_finish",
			X:      0,
			Y:      0,
			Color:  0,
		}, false)
	}

	runtime.GC()
}

// AddPixAction 添加像素
func (inst *PixelGameControlle) AddPixAction(sub *models.Subscriber, req interface{}) {
	reqData := req.(*netws.PixelAddReq)

	// 验证ws连接
	if _, ok := netws.GetSubscriber(sub); !ok {
		pushAuthMsg(sub, "auth error", "", "")
		netws.Leave(sub)
		return
	}

	uid := sub.Uid
	if uid == 0 {
		pushAuthMsg(sub, "auth error", "", "")
		netws.Leave(sub)
		return
	}

	// 速率控制
	if !sub.Limiter.Allow() {
		log.Info("", fmt.Sprintf("触发速率限制 uid:%d", uid))
		return
	}

	// 目前最大尺寸支持 1000*1000
	if reqData.X > 999 || reqData.Y > 999 || reqData.Color > 255 {
		log.Info("", fmt.Sprintf("违法数据 uid:%d x:%d y:%d c:%d", uid, reqData.X, reqData.Y, reqData.Color))
		return
	}

	// 保存画布数据
	service := services.PixelService{}
	requiredLogin, err := service.SavePixel(reqData.X, reqData.Y, reqData.Color, uid)
	if err == nil {
		// 广播数据
		pushBroadcast(reqData)
	}

	// 回复客户端
	pushPixelMsg(sub, reqData, requiredLogin)
}

// PixelChat 接受聊天内容
func (inst *PixelGameControlle) PixelChat(sub *models.Subscriber, req interface{}) {
	reqData := req.(*netws.PixelChatReq)

	// 验证ws连接
	if _, ok := netws.GetSubscriber(sub); !ok {
		pushAuthMsg(sub, "auth error", "", "")
		netws.Leave(sub)
		return
	}

	// 保存聊天记录
	if true {
		fmt.Println(reqData.Msg)
		pushChatMsg(sub, 1)
	}

	userInfo, err := models.GetGameUserById(sub.Uid)
	if err != nil {
		pushAuthMsg(sub, "10001", "", "")
		netws.Leave(sub)
		return
	}

	if userInfo == nil {
		pushAuthMsg(sub, "token error", "", "")
		netws.Leave(sub)
		return
	}

	// 获取当前时间
	currentTime := time.Now()

	// 格式化日期和时间
	dateString := currentTime.Format("2006/01/02 15:04")

	// 过滤XSS攻击
	p := bluemonday.UGCPolicy()
	safeInput := p.Sanitize(reqData.Msg)
	if len(safeInput) != 0 {
		// 过滤敏感词
		safeInput = sensitive.SensitiveFilter.Replace(safeInput, '*')

		nickname := fmt.Sprintf("%s %s", userInfo.Nickname, dateString)
		service := services.PixelService{}
		_ = service.SaveChat(nickname, safeInput)

		// 广播数据
		pushBroadcast(map[string]string{
			"action":   "chat_push",
			"msg":      safeInput,
			"msg_id":   sub.Uuid,
			"nickname": nickname,
		})
	}

}
