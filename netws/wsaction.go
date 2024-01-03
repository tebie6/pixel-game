package netws

import (
	"github.com/tebie6/pixel-game/models"
	"time"
)

var (
	// actAgentMap 是一个映射，用于存储与不同的消息代码对应的行为函数
	actAgentMap = map[string]func(*models.Subscriber, interface{}){
		WSActionPing: pingAct,
	}
)

// Router 注册长连接路由
func Router(inst WsActionInf) {
	actMap := inst.Init()
	for msgCode, f := range actMap {
		actAgentMap[msgCode] = f
	}
}

// pingAct 是处理 ping 命令的函数
func pingAct(sub *models.Subscriber, content interface{}) {
	// 设置订阅者的最后pong时间
	sub.Ping = time.Now().Unix()

	// 发布一个事件到消息队列
	PublishMSG(&models.EventModel{
		Type:      models.EventHeartbeat,
		Sub:       sub,
		Timestamp: sub.Ping,
		Content: &HeartBeatPong{
			Action:    WSActionPong,
			Timestamp: sub.Ping,
		},
	})
}
