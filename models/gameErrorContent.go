package models

// GameErrorContent 错误记录
type GameErrorContent struct {
	Id        int64  `gorm:"primary_key" json:"id"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	Lineno    string `json:"lineno"`
	Colno     string `json:"colno"`
	Stack     string `json:"stack"`
	Uid       int64  `json:"uid"`
	CreatedAt string `json:"created_at"`
}

// 自定义表名
func (*GameErrorContent) TableName() string {
	return "game_error_content"
}
