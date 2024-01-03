package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type IO struct {
	name string
	filepath string
	fp *os.File
	buffer chan string
	dateStr string
}

type Logger struct {
	basePath string
	members *sync.Map // map[string]*IO
	dateSplitFmt string
}

var (
	logger *Logger
)

func InitLogger(basepath string, dateFmt string) {

	if _, err := os.Stat(basepath); os.IsNotExist(err) {
		panic(fmt.Errorf("the dir %s that for logger is not exist", basepath))
	}

	logger = &Logger{
		basePath: strings.TrimRight(basepath, "/"),
		members:  new(sync.Map),
		dateSplitFmt: dateFmt ,
	}
}

func DestroyLogger() {
	fmt.Println("dl called")
	logger.members.Range(func(key, value interface{}) bool {
		prefix := key.(string)
		fmt.Println("prefix ", prefix, "will be destroy")
		logio := value.(*IO)
		close(logio.buffer)
		return true
	})
}

func WLog(prefix ,msg string) (err error) {
	val,ok := logger.members.Load(prefix)
	var io *IO
	if !ok {
		fmt.Println("init logger ", prefix)
		io = new(IO)
		io.name = prefix
		io.dateStr = time.Now().Format(logger.dateSplitFmt)
		io.filepath = fmt.Sprintf("%s/%s-%s.log", logger.basePath, prefix, io.dateStr)
		io.fp,err = os.OpenFile(io.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err!=nil {
			return err
		}
		io.buffer = make(chan string, 1000)

		logger.members.Store(prefix, io)

		go writerAwaiter(io)

	}else{
		io = val.(*IO)

	}

	io.buffer<-msg
	return nil
}

func writerAwaiter(io *IO) {
	//@todo deal with panic
	defer printPanicStackError()

	defer func(){
		err := io.fp.Close()
		if err!=nil {
			fmt.Println("Err: close file failed ", err)
		}

		logger.members.Delete(io.name)

		fmt.Println("clean logger ", io.name, " done")
	}()

	//var err error

	for {
		select {
		case msg,ok := <-io.buffer :
			if !ok {
				fmt.Println("recive close channel")
				return
			}

			fmt.Println("recv logger msg ", msg)

			msg = fmt.Sprintf("%s\n", msg)

			// ?? need split
			dateStr := time.Now().Format(logger.dateSplitFmt)
			if dateStr!=io.dateStr {
				err := io.fp.Close()
				if err!=nil {
					fmt.Println("Err: close file failed ", err)
				}

				io.dateStr = dateStr
				io.filepath = fmt.Sprintf("%s/%s-%s.log", logger.basePath, io.name, io.dateStr)
				io.fp,err = os.OpenFile(io.filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err!=nil {
					fmt.Println("Err: new log file failed ", err)
				}
			}

			_,err := io.fp.Write([]byte(msg))

			if err != nil {
				fmt.Println("Err: write log failed", err)
				break
			}
		}
	}
}