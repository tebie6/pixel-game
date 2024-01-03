/*
 @Title  数据模型主文件
 @Description  请填写文件描述（需要改）
 @Author  Leo  2020/4/21 2:55 下午
 @Update  Leo  2020/4/21 2:55 下午
*/

package models

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/tebie6/pixel-game/conf"
	"time"
)

var (
	// 全局连接句柄
	dbs     = make(map[string]*gorm.DB)
	redises = make(map[string]*redis.Pool)
)

func InitModel() {

	// db_main
	err := registerDB("db_main")
	if err != nil {
		panic(err)
	}

	// redis_ucenter
	err = registerRedis("redis_main")
	if err != nil {
		panic(err)
	}

}

// 获取主库实例
func GetDbInst() *gorm.DB {
	return dbs["db_main"]
}

// 获取redis
func GetRedisInst() *redis.Pool {
	return redises["redis_main"]
}

func registerDB(sectionName string) error {
	//"user:password@tcp(host:port)/dbname?charset=utf8&parseTime=True&loc=Local"
	connstr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", conf.GetConfigString(sectionName, "user"),
		conf.GetConfigString(sectionName, "password"),
		conf.GetConfigString(sectionName, "host"),
		conf.GetConfigString(sectionName, "port"),
		conf.GetConfigString(sectionName, "name"))

	var err error
	dbs[sectionName], err = gorm.Open("mysql", connstr)
	if err != nil {
		return err
	}

	// todo 设置连接池，实际设置参考正式上线使用量
	//db.DB().SetMaxIdleConns()

	return nil
}

func registerRedis(sectionName string) error {
	// 建立连接池
	redisHost := conf.GetConfigString(sectionName, "host")
	redisPort := conf.GetConfigString(sectionName, "port")
	redisPwd := conf.GetConfigString(sectionName, "password")
	redisDb, err := conf.GetConfigInt(sectionName, "db")

	if err != nil {
		redisDb = 0
	}

	redises[sectionName] = &redis.Pool{
		MaxIdle:     10,               // 最大空闲连接数，即会有这么多个连接提前等待着，但过了超时时间也会关闭
		MaxActive:   50,               // 最大连接数，即最多的tcp连接数，一般建议往大的配置，但不要超过操作系统文件句柄个数（centos下可以ulimit -n查看）
		IdleTimeout: 60 * time.Second, // 空闲连接超时时间，但应该设置比redis服务器超时时间短。否则服务端超时了，客户端保持着连接也没用
		Wait:        true,             // 如果超过最大连接，是报错，还是等待
		Dial: func() (redis.Conn, error) {
			con, err := redis.Dial("tcp", redisHost+":"+redisPort,
				redis.DialPassword(redisPwd),
				redis.DialDatabase(int(redisDb)),
				redis.DialConnectTimeout(2*time.Second),
				redis.DialReadTimeout(2*time.Second),
				redis.DialWriteTimeout(2*time.Second))
			if err != nil {
				return nil, err
			}
			return con, nil
		},
	}

	return nil
}
