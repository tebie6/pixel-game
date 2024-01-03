/*
 @Title  日志
 @Description
 @Author  Leo  2020/4/20 5:05 下午
 @Update  Leo  2020/4/20 5:05 下午
*/

package log

import (
	"fmt"
	"github.com/tebie6/pixel-game/tools"
	"github.com/tebie6/pixel-game/tools/logger"
	"runtime"
	"strings"
	"time"
)

const (
	ErrorLevelDebug = 0
	ErrorLevelInfo  = 1
	ErrorLevelWarn  = 2
	ErrorLevelError = 3
	ErrorLevelAlert = 4

	ErrorLevelNo = 100 // 不记录任何日志
)

var (
	ErrorLevel = 0
)

func SetLogErrorLevel(l int) {
	ErrorLevel = l
}

func PrintPanicStackError() {
	if x := recover(); x != nil {
		Error("errstack", "panic %+v", x)
		PrintPanicStack()
	}
}

func PrintPanicStack() {
	stack := make([]string, 0)
	for i := 0; i < 10; i++ {
		funcName, file, line, ok := runtime.Caller(i)
		if ok {
			funcName := runtime.FuncForPC(funcName).Name()
			//WriteLog(LogPanic,"frame %d:[func:%s, file: %s, line:%d]", i, funcName, file, line)
			stack = append(stack, fmt.Sprintf("frame %d:[func:%s, file: %s, line:%d]", i, funcName, file, line))
		}
	}

	Error("errstack", strings.Join(stack, "\n")+"---END---\n")
}

// Log an error msg
// prefix: if null string, it will be system
// level: the degree of the exception
// f: the format of Sprintf
// args: the args of Sprintf
func writelog(prefix, level, f string, args ...interface{}) {
	if prefix == "" {
		prefix = "system"
	}
	msg := fmt.Sprintf(f, args...)
	now := time.Now().Format("2006-01-02.15:04:05")
	caller := tools.GetCaller(2)
	msg = fmt.Sprintf("%s\t%s\t%s\t%s", now, level, caller, msg)
	err := logger.WLog(prefix, msg)
	if err != nil {
		fmt.Println("write log failed!!!")
	}
}

func Debug(p, f string, args ...interface{}) {
	if ErrorLevel > ErrorLevelDebug {
		return
	}
	writelog(p, "D", f, args...)
}

func Info(p, f string, args ...interface{}) {
	if ErrorLevel > ErrorLevelInfo {
		return
	}
	writelog(p, "I", f, args...)
}

func Warning(p, f string, args ...interface{}) {
	if ErrorLevel > ErrorLevelWarn {
		return
	}
	writelog(p, "W", f, args...)
}

func Error(p, f string, args ...interface{}) {
	if ErrorLevel > ErrorLevelError {
		return
	}
	writelog(p, "E", f, args...)
}

func Alert(p, f string, args ...interface{}) {
	if ErrorLevel > ErrorLevelAlert {
		return
	}
	writelog(p, "A", f, args...)
}
