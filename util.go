package logger

import (
	"fmt"
	"path"
	"runtime"
	"time"
)

// 日志队列数据结构体
type LogData struct {
	Message      string
	TimeStr      string
	LevelStr     string
	Filename     string
	FuncName     string
	LineNo       int
	WarnAndFatal bool // 指定是否写入到警告日志
}

func GetLineInfo() (fileName string, funcName string, lineNo int) {
	pc, file, line, ok := runtime.Caller(4)
	if ok {
		fileName = file
		funcName = runtime.FuncForPC(pc).Name()
		lineNo = line
	}
	return
}

/*
1. 当业务调用打日志的方法时，我们把日志相关的数据写入到chan（队列）
2. 然后我们有一个后台的线程不断的从chan里面获取这些日志，最终写入到文件。
*/
func writeLog(level int, format string, args ...interface{}) *LogData {
	nowStr := time.Now().Format(TIME_FORMAT)
	levelStr := getLevelText(level)

	fileName, funcName, lineNo := GetLineInfo()

	fileName = path.Base(fileName)
	funcName = path.Base(funcName)
	msg := fmt.Sprintf(format, args...)

	logData := &LogData{
		Message:      msg,
		TimeStr:      nowStr,
		LevelStr:     levelStr,
		Filename:     fileName,
		FuncName:     funcName,
		LineNo:       lineNo,
		WarnAndFatal: false,
	}

	// 判断是否需要写入到警告日志
	if level == LogLevelError || level == LogLevelWarn || level == LogLevelFatal {
		logData.WarnAndFatal = true
	}

	return logData
	//fmt.Fprintf(file, "%s %s (%s:%s:%d) %s\n", nowStr, levelStr, fileName, funcName, lineNo, msg)
}
