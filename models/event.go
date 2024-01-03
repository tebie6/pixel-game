/*
@Title  ws 相关数据模型
*/
package models

import (
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
)

type EventType int

const (
	// 事件定义
	_ = iota
	_
	EventHeartbeat
	EventMessage
	EventBroadcast
)

// ws 消息
type EventModel struct {
	Type      EventType   // JOIN, LEAVE, MESSAGE
	Sub       *Subscriber // 广播请传 nil
	Timestamp int64       // Unix timestamp (secs)
	Content   interface{}
}

// Subscriber 代表一个客户端连接
type Subscriber struct {
	Uuid string // 连接唯一ID
	Ping int64  // 最后心跳时间
	Uid  int64  // 用户ID

	Conn    *websocket.Conn // ws 客户端连接句柄
	Request *http.Request   // http 请求

	WriteChan     chan *EventModel // 发消息通道 (这里用指针是为了防止 channel 拷贝数据，造成资源浪费，特别是广播时候)
	WriteChanLock sync.Mutex
	StopChan      chan bool     // 用于监控连接中断信号
	Limiter       *rate.Limiter // 令牌桶 限制速率
}

// default 模型用，写入消息到写通道
func (sub *Subscriber) WriteMsg(event *EventModel) {

	sub.WriteChanLock.Lock()
	defer sub.WriteChanLock.Unlock()

	if len(sub.WriteChan) > 4 {
		// TODO 记录LOG
	}

	sub.WriteChan <- event
}
