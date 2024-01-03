package netws

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/tebie6/pixel-game/conf"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/tools"
	"github.com/tebie6/pixel-game/tools/log"
	"github.com/tebie6/pixel-game/tools/proc/wspriv"
	string2 "github.com/tebie6/pixel-game/tools/string"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func defaultWSLoop() {

	r := gin.New()

	r.GET("/connect", defaultWSLoopConnectWS)

	wsListenAddress := fmt.Sprintf("%s:%s",
		conf.GetConfigString("ws", "host"),
		conf.GetConfigString("ws", "port"),
	)
	err := r.Run(wsListenAddress)
	if err != nil {
		log.Error("", "start server failed %s", err.Error())
	}
}

func defaultWSLoopConnectWS(c *gin.Context) {

	// 升级协议
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// TODO 记录日志
		return
	}

	uuidObj := uuid.NewV4()
	connUuid := uuidObj.String()

	wsWriteConn := make(chan *models.EventModel, 100)

	// 初始化连接句柄
	sub := &models.Subscriber{
		Uuid:      connUuid,
		Conn:      ws,
		Request:   c.Request,
		WriteChan: wsWriteConn,
		StopChan:  make(chan bool),
	}

	msgList := tools.NewPipe()

	// 加入连接
	Join(sub)

	// 断开连接
	defer Leave(sub)

	// 写缓冲协程
	go startWriteListener(sub, msgList)

	// 写协程
	go startWriteLoop(sub, msgList)

	// 关闭回收协程
	go closeGoroutine(sub, msgList)

	// 接收ws消息
	for {
		var messageType int
		var raw []byte

		// 读取数据
		messageType, raw, err = ws.ReadMessage()
		if err != nil {
			log.Info("", "退出接收ws消息\n", err.Error())
			return
		}

		switch messageType {
		case websocket.TextMessage:

			// 通过action获取定义的方法
			msgID, msg, err := wspriv.DecodeMsg(raw)
			if err == nil {
				// 判断ws路由是否存在
				actFunc, ok := actAgentMap[msgID]
				if ok {
					actFunc(sub, msg)
				}
			} else {
				log.Info("", "recv user msg, but failed to decode", err.Error())
			}
		}
	}
}

// 写缓存协程
func startWriteListener(sub *models.Subscriber, l *tools.Pipe) {

	// 从channel中取数据
	for {
		select {
		case event, ok := <-sub.WriteChan:
			if !ok {
				fmt.Println("startWriteListener end")
				return
			}

			l.Add(event.Content)
		}
	}
}

// 写协程
func startWriteLoop(sub *models.Subscriber, l *tools.Pipe) {

	var err error
	var data []interface{}
	for {
		data = data[0:0] // 初始化切片

		// 从写缓存list获取数据
		exit := l.Pick(&data)
		if exit {
			fmt.Println("startWriteLoop end")
			return
		}

		// 如果消息积压
		if len(data) > 4 {
			log.Info("", "%s WriterList %d overstock", sub.Uuid, len(data))
		}

		for _, msg := range data {

			var result []byte
			result, err = json.Marshal(msg)
			if err != nil {
				log.Error("", "pack msg failed %s", err.Error())
				return
			}

			// 压缩字符串
			result, _ = string2.CompressString(string(result))

			// 给客户端推送消息
			err = sub.Conn.WriteMessage(websocket.BinaryMessage, result)
			if err != nil {
				log.Warning("", "write msg failed %s", err.Error())
			}
		}
	}
}

func closeGoroutine(sub *models.Subscriber, l *tools.Pipe) {
	for {
		select {
		case <-sub.StopChan:
			// 最大可能保证数据发送出去，防止数据未发送完程序终止
			time.Sleep(5 * time.Second)
			// 回收写缓冲协程
			close(sub.WriteChan)
			// 回收写协程
			l.Add(nil)
			fmt.Println("StopChan closeGoroutine")
			return
		}
	}
}

// Deal messages for WebSocket users.
func eventHandleDefault(event *models.EventModel) {
	switch event.Type {
	case models.EventHeartbeat:
		if nil == event.Sub {
			log.Error("", "Heartbeat send failed: illegal subscriber")
			break
		}
		event.Sub.WriteMsg(event)
		break
	case models.EventMessage:
		// 点对点
		if nil == event.Sub {
			log.Error("", "Peer to peer send failed: illegal subscriber")
			break
		}
		event.Sub.WriteMsg(event)
		break
	case models.EventBroadcast:
		subscribers.Range(func(k interface{}, v interface{}) bool {
			sub, ok := v.(*models.Subscriber)
			if ok && sub != nil {
				sub.WriteMsg(event)
			}
			return true
		})
		break
	}
}
