package logger

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// 定义文件日志结构体实现日志接口
type FileLogger struct {
	level         int
	logPath       string
	logName       string
	file          *os.File
	warnFile      *os.File
	LogDataChan   chan *LogData // 队列存放写入的日志，防止值拷贝，传入LogData指针
	logSplitType  int           // 日志切分方式
	logSplitSize  int64         // 日志切分大小
	lastSplitHour int           // 上次切分日志的小时
}

// 构造函数，返回一个LogInterface类型的实例或者错误
func NewFileLogger(config map[string]string) (log LogInterface, err error) {
	logPath, ok := config["log_path"]
	if !ok {
		err = fmt.Errorf("not found log_path ")
		return
	}

	logName, ok := config["log_name"]
	if !ok {
		err = fmt.Errorf("not found log_name ")
		return
	}

	logLevel, ok := config["log_level"]
	if !ok {
		err = fmt.Errorf("not found log_level ")
		return
	}

	logChanSize, ok := config["log_chan_size"]
	if !ok {
		logChanSize = "50000"
	}

	var logSplitType = LogSplitTypeHour
	var logSplitSize int64
	logSplitStr, ok := config["log_split_type"]
	if !ok {
		logSplitStr = "hour"  // 日志按小时切分
	} else {
		if logSplitStr == "size" { // 日志按文件大小切分
			logSplitSizeStr, ok := config["log_split_size"]
			if !ok {
				logSplitSizeStr = "104857600" // 默认100M
			}

			logSplitSize, err = strconv.ParseInt(logSplitSizeStr, 10, 64)
			if err != nil {
				logSplitSize = 104857600
			}

			logSplitType = LogSplitTypeSize
		} else {
			logSplitType = LogSplitTypeHour
		}
	}

	chanSize, err := strconv.Atoi(logChanSize)
	if err != nil {
		chanSize = 50000
	}

	level := getLogLevel(logLevel)
	log = &FileLogger{
		level:         level,
		logPath:       logPath,
		logName:       logName,
		LogDataChan:   make(chan *LogData, chanSize), // 初始化LogData队列
		logSplitSize:  logSplitSize,
		logSplitType:  logSplitType,
		lastSplitHour: time.Now().Hour(),
	}
	log.Init()
	//fmt.Printf("logSplitTpye: %s\nlogSpliteSize: %d\nlogSplitStr: %s\n", logSplitType, logSplitSize, logSplitStr)
	return
}

func (f *FileLogger) Init() {
	filename := fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open faile %s failed, err:%v", filename, err))
	}

	f.file = file

	//写错误日志和fatal日志的文件
	filename = fmt.Sprintf("%s/%s.log.wf", f.logPath, f.logName)
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open faile %s failed, err:%v", filename, err))
	}

	f.warnFile = file
	go f.writeLogBackground() // 启动一个线程，将日志队列中的日志写入文件
}

// 日志按小时切分的函数实现
func (f *FileLogger) splitFileHour(warnFile bool) {
	now := time.Now()
	hour := now.Hour()
	if hour == f.lastSplitHour {
		return
	}

	f.lastSplitHour = hour
	var backupFilename string
	var filename string

	if warnFile {
		backupFilename = fmt.Sprintf("%s/%s.log.wf_%04d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)

		filename = fmt.Sprintf("%s/%s.log.wf", f.logPath, f.logName)
	} else {
		backupFilename = fmt.Sprintf("%s/%s.log_%04d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)
		filename = fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	}

	file := f.file
	if warnFile {
		file = f.warnFile
	}

	file.Close()
	os.Rename(filename, backupFilename) // 备份日志文件

	// 重新打开文件，写入新的日志
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}

	if warnFile {
		f.warnFile = file
	} else {
		f.file = file
	}
}

// 日志按大小切分的函数实现
func (f *FileLogger) splitFileSize(warnFile bool) {

	file := f.file
	if warnFile {
		file = f.warnFile
	}

	statInfo, err := file.Stat() // 获取文件信息
	if err != nil {
		return
	}

	fileSize := statInfo.Size()
	if fileSize <= f.logSplitSize { // 未达到指定文件大小
		return
	}
	// 达到指定文件大小
	var backupFilename string
	var filename string

	now := time.Now()
	if warnFile {
		backupFilename = fmt.Sprintf("%s/%s.log.wf_%04d%02d%02d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

		filename = fmt.Sprintf("%s/%s.log.wf", f.logPath, f.logName)
	} else {
		backupFilename = fmt.Sprintf("%s/%s.log_%04d%02d%02d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		filename = fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	}

	file.Close()
	os.Rename(filename, backupFilename) // 备份文件

	// 重新打开新的文件，写入新的日志
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}

	if warnFile {
		f.warnFile = file
	} else {
		f.file = file
	}
}

func (f *FileLogger) checkSplitFile(warnFile bool) {
	// 按小时切分日志
	if f.logSplitType == LogSplitTypeHour {
		f.splitFileHour(warnFile)
		return
	}
	// 按照文件大小切分日志
	f.splitFileSize(warnFile)
}

// 异步处理，将日志从队列中取出，写入到文件
func (f *FileLogger) writeLogBackground() {
	for logData := range f.LogDataChan {
		var file = f.file
		if logData.WarnAndFatal {
			file = f.warnFile
		}

		f.checkSplitFile(logData.WarnAndFatal)

		fmt.Fprintf(file, "%s %s (%s:%s:%d) %s\n", logData.TimeStr,
			logData.LevelStr, logData.Filename, logData.FuncName, logData.LineNo, logData.Message)
	}
}

func (f *FileLogger) SetLevel(level int) {
	if level < LogLevelDebug || level > LogLevelFatal {
		level = LogLevelDebug
	}
	f.level = level
}

func (f *FileLogger) Debug(format string, args ...interface{}) {
	if f.level > LogLevelDebug {
		return
	}

	logData := writeLog(LogLevelDebug, format, args...)
	select { // 判断队列是否已经满了
	case f.LogDataChan <- logData: // 队列没有满，正常写入日志数据
	default: // 队列满了，丢弃当前这条日志，防止线程阻塞
	}
}

func (f *FileLogger) Trace(format string, args ...interface{}) {
	if f.level > LogLevelTrace {
		return
	}

	logData := writeLog(LogLevelTrace, format, args...)
	select { // 判断队列是否已经满了
	case f.LogDataChan <- logData: // 队列没有满，正常写入日志数据
	default: // 队列满了，丢弃当前这条日志，防止线程阻塞
	}
}

func (f *FileLogger) Info(format string, args ...interface{}) {
	if f.level > LogLevelInfo {
		return
	}

	logData := writeLog(LogLevelInfo, format, args...)
	select { // 判断队列是否已经满了
	case f.LogDataChan <- logData: // 队列没有满，正常写入日志数据
	default: // 队列满了，丢弃当前这条日志，防止线程阻塞
	}
}

func (f *FileLogger) Warn(format string, args ...interface{}) {
	if f.level > LogLevelWarn {
		return
	}

	logData := writeLog(LogLevelWarn, format, args...)
	select { // 判断队列是否已经满了
	case f.LogDataChan <- logData: // 队列没有满，正常写入日志数据
	default: // 队列满了，丢弃当前这条日志，防止线程阻塞
	}
}

func (f *FileLogger) Error(format string, args ...interface{}) {
	if f.level > LogLevelError {
		return
	}

	logData := writeLog(LogLevelError, format, args...)
	select { // 判断队列是否已经满了
	case f.LogDataChan <- logData: // 队列没有满，正常写入日志数据
	default: // 队列满了，丢弃当前这条日志，防止线程阻塞
	}
}

func (f *FileLogger) Fatal(format string, args ...interface{}) {
	if f.level > LogLevelFatal {
		return
	}

	logData := writeLog(LogLevelFatal, format, args...)
	select { // 判断队列是否已经满了
	case f.LogDataChan <- logData: // 队列没有满，正常写入日志数据
	default: // 队列满了，丢弃当前这条日志，防止线程阻塞
	}
}

func (f *FileLogger) Close() {
	f.file.Close()
	f.warnFile.Close()
}
