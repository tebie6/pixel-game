package netws

import (
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/tools/proc/wspriv"
	"reflect"
)

const (
	// 心跳
	WSActionPing = "ping"
	WSActionPong = "pong"

	// 权限验证
	WSActionAuthReq  = "auth"
	WSActionAuthResp = "auth_resp"

	// 像素
	WSActionPixelAddReq = "pixel_add"
	WSActionPixelResp   = "pixel_resp"

	// 像素聊天
	WSActionPixelChatReq  = "pixel_chat"
	WSActionPixelChatResp = "pixel_chat_resp"

	// 广播
	WSActionBroadcast = "broadcast"
)

// ping/pong
type HeartBeatPing struct {
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
}

type HeartBeatPong struct {
	Action    string `json:"action"`
	Timestamp int64  `json:"timestamp"`
}

// 权限验证
type AuthReq struct {
	Action   string `json:"action"`
	Token    string `json:"token"`
	Size     int    `json:"size"`
	LoadMode int    `json:"load_mode"`
}

type AuthResp struct {
	Action   string `json:"action"`
	Msg      string `json:"msg"`
	Nickname string `json:"nickname"`
	Id       string `json:"id"`
}

// 像素
type PixelAddReq struct {
	Action string `json:"action"`
	X      int64  `json:"x"`
	Y      int64  `json:"y"`
	Color  int64  `json:"color"`
}

type PixelResp struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type PixelPush struct {
	Data     interface{} `json:"d"`
	Progress string      `json:"p"`
}

// 像素聊天
type PixelChatReq struct {
	Action string `json:"action"`
	Msg    string `json:"msg"`
	MsgId  string `json:"msg_id"`
}

type PixelChatResp struct {
	Action string `json:"action"`
	Status int64  `json:"status"`
}

// 广播
type Broadcast struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type WsActionInf interface {
	Init() map[string]func(*models.Subscriber, interface{})
}

// 将结构体绑定到消息
func registerMetaMessages() {

	// 普通 json 文本消息
	m := map[string]interface{}{

		// 心跳
		WSActionPing: (*HeartBeatPing)(nil),
		WSActionPong: (*HeartBeatPong)(nil),

		// 验证
		WSActionAuthReq:  (*AuthReq)(nil),
		WSActionAuthResp: (*AuthResp)(nil),

		// 像素
		WSActionPixelAddReq: (*PixelAddReq)(nil),
		WSActionPixelResp:   (*PixelResp)(nil),

		// 像素聊天
		WSActionPixelChatReq:  (*PixelChatReq)(nil),
		WSActionPixelChatResp: (*PixelChatResp)(nil),

		// 广播
		WSActionBroadcast: (*Broadcast)(nil),
	}

	for k, v := range m {
		wspriv.AddProtoMsg(k, reflect.TypeOf(v).Elem())
	}
}
