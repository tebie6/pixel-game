/*
 @Title
 @Description
 @Author  Leo
 @Update  2020/8/17 5:53 下午
*/

package logger

import (
	"fmt"
	"runtime"
)

func printPanicStackError() {
	if x := recover(); x != nil {
		fmt.Println("panic ",x)
		printPanicStack()
	}
}

func printPanicStack() {
	for i := 0; i < 10; i++ {
		funcName, file, line, ok := runtime.Caller(i)
		if ok {
			funcName := runtime.FuncForPC(funcName).Name()
			errInfo := fmt.Sprintf("frame %d:[func:%s, file: %s, line:%d]", i, funcName, file, line)
			fmt.Println(errInfo)
		}
	}
}
