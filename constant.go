package logger

// 通过iota和常量定义日志级别
const (
	LogLevelDebug = iota
	LogLevelTrace
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

const (
	TIME_FORMAT      = "2006-01-02 15:04:05.999"
	LogSplitTypeHour = iota
	LogSplitTypeSize
)

// 通过int类型的日志级别获取对应的字符串
func getLevelText(level int) string {
	switch level {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	}
	return "UNKNOWN"
}

// 通过字符串形式的日志级别获取int类型的日志级别
func getLogLevel(level string) int {
	switch level {
	case "debug":
		return LogLevelDebug
	case "trace":
		return LogLevelTrace
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	case "fatal":
		return LogLevelFatal
	}
	return LogLevelDebug
}
