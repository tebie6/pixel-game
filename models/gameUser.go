package models

import (
	"github.com/jinzhu/gorm"
	"github.com/tebie6/pixel-game/tools/nickname"
	"time"
)

// GameUser 游戏用户
type GameUser struct {
	Id          int64  `gorm:"primary_key" json:"id"`
	Nickname    string `json:"nickname"`
	AccessToken string `json:"access_token"`
	Status      int8   `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// 自定义表名
func (*GameUser) TableName() string {
	return "game_user"
}

func CreateGameUser(user *GameUser) (int64, error) {
	db := GetDbInst()
	user.Nickname = nickname.GetNickname()
	user.Status = 1
	user.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	user.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")

	result := db.Create(&user)
	return user.Id, result.Error
}

func UpdateGameUser(user *GameUser) error {
	db := GetDbInst()
	user.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	return db.Save(&user).Error
}

func GetGameUserById(id int64) (*GameUser, error) {
	db := GetDbInst()
	user := new(GameUser)
	err := db.Where("id = ?", id).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, err
}

func GetUserByAccessToken(accessToken string) (*GameUser, error) {
	db := GetDbInst()
	user := new(GameUser)
	err := db.Where("access_token = ?", accessToken).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, err
}
