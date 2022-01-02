## logger日志库

### 1、安装
```bash
go get github.com/staryjie/logger@latest
```

### 2、使用

示例：
```go
package main

import (
	"github.com/staryjie/logger"
	"time"
)

func initLogger(name, logPath, logName, level, split_type string) (err error) {
	config := make(map[string]string)
	config["name"] = name
	config["log_path"] = logPath
	config["log_name"] = logName
	config["log_level"] = level
	config["log_split_type"] = split_type
	err = logger.InitLogger(name, config)
	if err != nil {
		return
	}
	logger.Trace("init logger success!")
	return
}

func Run() {
	for {
		logger.Fatal("User server is running...")
		time.Sleep(1 * time.Second)
	}

}

func main() {
	initLogger("file", "/var/log/", "user", "debug", "size")
	Run()
}
```

### 3、参数说明

```bash
name,           [file|console]  file: 写入文件, console: 输出日志到终端
log_path,       [str] 当`name`为`file`时需要传入，日志存放的路径
log_name,       [str] 当`name`为`file`时需要传入，日志文件名(不需要传入扩展名)
log_level,      [str] 日志级别，分别有 debug, trace, info, warn, error, fatal等
log_split_type, [str] 日志切割方式，默认是(file)按小时切割，如果需要按指定大小切割需要传入`size`和另一个参数`log_split_size`
log_split_size, [str] 日志切割大小，当日志文件大于等于该参数指定的值(B)时，进行切割，默认是100M，也就是 [104857600]
```

### 4、实现了哪些功能

- 通过Go的channel将实时产生的日志写入到一个channel，然后另起一个线程从channel中读取日志消息，写入到文件中
- 通过按小时进行日志文件切割或者指定日志大小进行文件切割