package services

import (
	"github.com/gomodule/redigo/redis"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/netws"
)

type UserService struct{}

// GetOnlineUserList 获取在线用户列表
func (r *UserService) GetOnlineUserList() []map[string]string {
	var res = make([]map[string]string, 0)
	for _, v := range netws.GetSubscriberList() {
		sub := v.(*models.Subscriber)
		if sub.Uid != 0 {
			nickname, _ := r.GetNicknameById(sub.Uid)
			res = append(res, map[string]string{
				"id":       sub.Uuid,
				"nickname": nickname,
			})
		}
	}

	return res
}

// GetNicknameById 通过uid获取昵称
func (r *UserService) GetNicknameById(uid int64) (nickname string, err error) {
	rdb := models.GetRedisInst()
	conn, _ := rdb.Dial()

	nickname, err = redis.String(conn.Do("HGET", "nickname", uid))
	if err != nil && err != redis.ErrNil {
		return "", err
	}

	if len(nickname) == 0 {
		userInfo, _ := models.GetGameUserById(uid)
		if userInfo == nil {
			return "", err
		}

		// 存储到缓存
		conn.Do("HSET", "nickname", uid, userInfo.Nickname)

		nickname = userInfo.Nickname
	}

	return nickname, nil
}
