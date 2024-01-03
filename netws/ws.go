package netws

import (
	"github.com/tebie6/pixel-game/models"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	subscribers sync.Map // map[string]*Subscriber ：连接集

	ConnectAmount int64 = 0 // 总连接数

	UserWsConnectionCounter *models.UserWsConnectionCounter
)

// 加入连接集合
func Join(sub *models.Subscriber) {
	sub.Ping = time.Now().Unix()
	subscribers.Store(sub.Uuid, sub)

	atomic.AddInt64(&ConnectAmount, 1)
}

// 连接断开
func Leave(sub *models.Subscriber) {
	// 防止连接不存在时重复执行，导致在线人数错误
	if _, ok := GetSubscriber(sub); !ok {
		return
	}

	// 暂停goroutine，让其释放
	close(sub.StopChan)

	subscribers.Delete(sub.Uuid)

	// 总连接数-1
	atomic.AddInt64(&ConnectAmount, -1)

	// 记录该用户总连接量
	UserWsConnectionCounter.RemoveConnection(sub.Uid)

	// 广播数据 - 当前人数
	event := &models.EventModel{
		Type:      models.EventBroadcast,
		Sub:       nil,
		Timestamp: time.Now().Unix(),
		Content: &Broadcast{
			Action: WSActionBroadcast,
			Data: map[string]string{
				"action": "online",
				"online": strconv.FormatInt(ConnectAmount, 10),
				"id":     sub.Uuid,
				"type":   "leave",
			},
		},
	}
	PublishMSG(event)

	defer func() {
		_ = sub.Conn.Close()
	}()

	// 主动调用GC回收内存
	runtime.GC()
}

// 获取连接信息
func GetSubscriber(sub *models.Subscriber) (interface{}, bool) {
	return subscribers.Load(sub.Uuid)
}

// 获取所有连接信息
func GetSubscriberList() []interface{} {
	var subscriberList = make([]interface{}, 0)
	subscribers.Range(func(k, v interface{}) bool {
		subscriberList = append(subscriberList, v)
		return true
	})

	return subscriberList
}

func InitWS() {

	// 注册action信息
	registerMetaMessages()

	// 启动ws服务
	RunWS()

	UserWsConnectionCounter = models.NewUserWsConnectionCounter()
}

func RunWS() {
	go defaultWSLoop()
}

// 推送一条消息
func PublishMSG(event *models.EventModel) {
	eventHandleDefault(event)
}
