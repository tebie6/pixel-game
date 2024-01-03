package wspriv

import (
	"encoding/json"
	"github.com/davyxu/cellnet"
	"reflect"
)

// BaseProto 是基础协议结构体，用于所有WebSocket消息
type BaseProto struct {
	Action string `json:"action"` // Action字段指定消息的类型或行为
}

// protoMap 和 protoMapByType 存储协议名称与协议类型之间的映射关系
var (
	protoMap       = make(map[string]reflect.Type)
	protoMapByType = make(map[reflect.Type]string)
)

// AddProtoMsg 函数用于将新的协议类型添加到映射表中
func AddProtoMsg(name string, t reflect.Type) {
	protoMap[name] = t
	protoMapByType[t] = name
}

// DecodeMsg 函数解码原始消息
func DecodeMsg(raw []byte) (msgId string, msg interface{}, err error) {
	base := new(BaseProto)
	err = json.Unmarshal(raw, base)
	if err != nil {
		return "", nil, err
	}
	ptype, ok := protoMap[base.Action]
	if !ok {
		return "", nil, cellnet.NewErrorContext("msg not exists", base.Action)
	}

	// 根据ptype创建消息实例并解码
	result := reflect.New(ptype).Interface()
	err = json.Unmarshal(raw, result)
	if err != nil {
		return "", nil, err
	}

	return base.Action, result, nil
}
