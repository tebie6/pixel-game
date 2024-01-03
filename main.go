package main

import (
	"fmt"
	"github.com/tebie6/pixel-game/conf"
	"github.com/tebie6/pixel-game/cron"
	"github.com/tebie6/pixel-game/models"
	"github.com/tebie6/pixel-game/netws"
	"github.com/tebie6/pixel-game/rpc"
	"github.com/tebie6/pixel-game/socketctrls"
	"github.com/tebie6/pixel-game/tools/log"
	"github.com/tebie6/pixel-game/tools/logger"
	"github.com/tebie6/pixel-game/tools/sensitive"
	"os"
)

func main() {

	// 初始化配置
	confpath := "./conf/conf.ini"
	err := conf.ParseConfigINI(confpath)
	if err != nil {
		fmt.Println("err : parse config failed", err.Error())
		os.Exit(1)
	}

	// 初始化log
	logPath := conf.GetConfigString("app", "log_path")

	logger.InitLogger(logPath, "20060102") // 日志
	el, err := conf.GetConfigInt("log", "level")
	if err != nil {
		el = 0
	}
	log.SetLogErrorLevel(int(el))

	// 初始化敏感词库
	sensitive.InitSensitive("./conf/sensitiveDict.txt")

	// 初始化数据库
	models.InitModel()

	// 初始化cron
	cron.InitCron()

	// 初始化ws
	netws.InitWS()
	netws.Router(&socketctrls.PixelGameControlle{})

	// 初始化RPC
	rpc.InitRpc()
}
