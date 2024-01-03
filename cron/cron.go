package cron

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/tebie6/pixel-game/services"
	"github.com/tebie6/pixel-game/tools/log"
	"time"
)

var pixelService = services.PixelService{} // 全局实例

// InitCron 初始化定时任务
func InitCron() {
	c := cron.New(cron.WithSeconds()) // 使用 WithSeconds 选项以支持到秒的精度

	// 每秒执行一次
	c.AddFunc("@every 5s", func() {
		err := pixelService.GenerateCanvasImage()
		if err != nil {
			log.Error("cron", fmt.Sprintf("Error executing GenerateCanvasImage: %v", err))
		} else {
			log.Info("cron", fmt.Sprintf("任务执行成功：%s", time.Now()))
		}
	})

	c.Start()
}
