package tools

import "sync"

// Pipe 结构体表示一个线程安全的数据管道
type Pipe struct {
	list      []interface{} // 存储数据的切片
	listGuard sync.Mutex    // 互斥锁，用于保护对 list 的访问
	listCond  *sync.Cond    // 条件变量，用于协调对 list 的访问
}

// NewPipe 创建并返回一个新的 Pipe 实例
func NewPipe() *Pipe {
	self := &Pipe{}
	self.listCond = sync.NewCond(&self.listGuard) // 初始化条件变量。
	return self
}

// Add 方法向 Pipe 中添加一个消息
func (p *Pipe) Add(msg interface{}) {
	p.listGuard.Lock()           // 加锁以保护 list。
	p.list = append(p.list, msg) // 将消息追加到 list 中
	p.listGuard.Unlock()         // 解锁
	p.listCond.Signal()          // 通知等待 list 变更的 p.listCond.Wait()
}

// Reset 方法重置 Pipe 的内部列表
func (p *Pipe) Reset() {
	p.list = p.list[0:0] // 清空 list。
}

// Pick 方法从 Pipe 中提取数据
func (p *Pipe) Pick(retList *[]interface{}) (exit bool) {

	p.listGuard.Lock()

	for len(p.list) == 0 {
		p.listCond.Wait() // 如果 list 为空，则等待。 Wait()调用时会自动解锁，唤醒时会自动加锁
	}

	p.listGuard.Unlock()

	p.listGuard.Lock()

	for _, data := range p.list {

		if data == nil {
			exit = true // 如果数据为 nil，则设置退出标志并中断循环。
			break
		} else {
			*retList = append(*retList, data) // 否则，将数据追加到 retList 中
		}
	}

	p.Reset()            // 重置 Pipe
	p.listGuard.Unlock() // 解锁
	return               // 返回退出标志
}
