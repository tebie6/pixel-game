package string

import (
	"bytes"
	"compress/gzip"
	"math/rand"
	"sync"
	"time"
)

func RandStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
	for i := 0; i < length; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

// gzipWriterPool 是一个用于重用 gzip.Writer 实例的 sync.Pool
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		// 创建一个新的 gzip.Writer 实例
		return gzip.NewWriter(nil)
	},
}

// CompressString 函数接收一个字符串并返回它的 gzip 压缩版本
func CompressString(str string) ([]byte, error) {
	var buf bytes.Buffer

	// 从池中获取一个 gzip.Writer 实例
	gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
	// 重置 gzip.Writer 以写入新的缓冲区
	gzipWriter.Reset(&buf)

	// 写入要压缩的字符串
	_, err := gzipWriter.Write([]byte(str))
	if err != nil {
		gzipWriter.Close()
		gzipWriterPool.Put(gzipWriter)
		return nil, err
	}

	// 关闭 gzip.Writer
	if err := gzipWriter.Close(); err != nil {
		gzipWriterPool.Put(gzipWriter)
		return nil, err
	}

	// 将 gzip.Writer 放回池中以供再次使用
	gzipWriterPool.Put(gzipWriter)
	// 返回压缩后的数据
	return buf.Bytes(), nil
}
