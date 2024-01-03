package models

// GamePixelRecord 像素操作记录
type GamePixelRecord struct {
	Id        int64  `gorm:"primary_key" json:"id"`
	X         int64  `json:"x"`
	Y         int64  `json:"y"`
	Color     int64  `json:"color"`
	Uid       int64  `json:"uid"`
	CreatedAt string `json:"created_at"`
}

// 自定义表名
func (*GamePixelRecord) TableName() string {
	return "game_pixel_record"
}
