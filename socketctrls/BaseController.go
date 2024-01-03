package socketctrls

import (
	"fmt"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/netws"
	"time"
)

// pushAuthMsg 推送验证消息
func pushAuthMsg(sub *models.Subscriber, msg string, nickname string, id string) {
	event := &models.EventModel{
		Type:      models.EventMessage,
		Sub:       sub,
		Timestamp: time.Now().Unix(),
		Content: &netws.AuthResp{
			Action:   netws.WSActionAuthResp,
			Msg:      msg,
			Nickname: nickname,
			Id:       id,
		},
	}

	netws.PublishMSG(event)
}

// pushPixelMsg 推送像素消息
func pushPixelMsg(sub *models.Subscriber, data interface{}, requiredLogin bool) {
	event := &models.EventModel{
		Type:      models.EventMessage,
		Sub:       sub,
		Timestamp: time.Now().Unix(),
		Content: &netws.PixelResp{
			Action: netws.WSActionPixelResp,
			Data:   data,
		},
	}

	netws.PublishMSG(event)
}

// pushPixelMsgUpdateCanvas 推送像素消息-更新画布
func pushPixelMsgUpdateCanvas(sub *models.Subscriber, data interface{}, pushTotal int, total int) {
	pushPixelMsgData := netws.PixelPush{
		Data:     data,
		Progress: fmt.Sprintf("%.2f", float64(pushTotal)/float64(total)*100),
	}
	pushPixelMsg(sub, pushPixelMsgData, false)
}

// pushChatMsg 推送聊天消息
func pushChatMsg(sub *models.Subscriber, status int64) {
	event := &models.EventModel{
		Type:      models.EventMessage,
		Sub:       sub,
		Timestamp: time.Now().Unix(),
		Content: &netws.PixelChatResp{
			Action: netws.WSActionPixelChatResp,
			Status: status,
		},
	}

	netws.PublishMSG(event)
}

// pushBroadcast 推送广播
func pushBroadcast(data interface{}) {
	event := &models.EventModel{
		Type:      models.EventBroadcast,
		Sub:       nil,
		Timestamp: time.Now().Unix(),
		Content: &netws.Broadcast{
			Action: netws.WSActionBroadcast,
			Data:   data,
		},
	}

	netws.PublishMSG(event)
}
