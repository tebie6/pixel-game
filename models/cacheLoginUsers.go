package models

import "sync"

// UserWsConnectionCounter 用于记录 WebSocket 连接数
type UserWsConnectionCounter struct {
	mu     sync.RWMutex
	counts map[int64]int64
}

// NewUserWsConnectionCounter 创建一个新的 UserWsConnectionCounter 实例
func NewUserWsConnectionCounter() *UserWsConnectionCounter {
	return &UserWsConnectionCounter{
		counts: make(map[int64]int64),
	}
}

// AddConnection 为指定的 UID 增加一个连接
func (c *UserWsConnectionCounter) AddConnection(uid int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[uid]++
}

// RemoveConnection 为指定的 UID 移除一个连接
func (c *UserWsConnectionCounter) RemoveConnection(uid int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[uid]--
	if c.counts[uid] <= 0 {
		delete(c.counts, uid)
	}
}

// GetConnectionCount 获取指定 UID 的连接数
func (c *UserWsConnectionCounter) GetConnectionCount(uid int64) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.counts[uid]
}
