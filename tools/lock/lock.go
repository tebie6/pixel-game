package lock

import (
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

// ObtainLock 获得锁
func ObtainLock(conn redis.Conn, lockKey string, timeout time.Duration) bool {
	reply, err := redis.String(conn.Do("SET", lockKey, "1", "EX", int(timeout.Seconds()), "NX"))
	if err != nil && err != redis.ErrNil {
		log.Fatalf("Error while trying to obtain lock: %v", err)
	}
	return reply == "OK"
}

// ReleaseLock 释放锁
func ReleaseLock(conn redis.Conn, lockKey string) {
	_, err := conn.Do("DEL", lockKey)
	if err != nil {
		log.Fatalf("Error while releasing lock: %v", err)
	}
}
