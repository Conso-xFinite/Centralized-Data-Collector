package logger

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	debugLogger = logrus.New()
	infoLogger  = logrus.New()
	warnLogger  = logrus.New()
	errorLogger = logrus.New()
	fatalLogger = logrus.New()

	logDir           string
	maxAgeDays       int
	lastDate         atomic.Value // string
	mu               sync.Mutex
	debugEnabled     bool = false // 调试开关, 默认关闭
	writeFileEnabled bool = false // 是否写文件, 默认关闭
)

func InitLoggers(dir string, age int) {
	logDir = dir
	maxAgeDays = age
	date := time.Now().Format("2006-01-02")
	lastDate.Store(date)
	updateLoggerFiles(date)
}

type SimpleFormatter struct{}

func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// level := strings.ToUpper(entry.Level.String())
	// // 标签统一宽度7（如 [DEBUG ]、[INFO  ]）
	// var label string
	// switch level {
	// case "DEBUG":
	// 	label = "[DEBUG] "
	// case "INFO":
	// 	label = "[INFO]  "
	// case "WARNING":
	// 	label = "[WARN]  "
	// case "ERROR":
	// 	label = "[ERROR] "
	// case "FATAL":
	// 	label = "[FATAL] "
	// default:
	// 	label = fmt.Sprintf("[%-7s]", level)
	// }
	// timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	// msg := fmt.Sprintf("%s %s %s\n", timestamp, label, entry.Message)
	// return []byte(msg), nil
	msg := fmt.Sprintf("%s\n", entry.Message)
	return []byte(msg), nil
}

const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

func _format(level logrus.Level, msg string) string {
	var levelLabel string
	var colorPrefix, colorBoldPrefix, colorSuffix string = "", "", ""
	switch level {
	case logrus.DebugLevel:
		colorPrefix = Blue
		colorBoldPrefix = Blue
		colorSuffix = Reset
		levelLabel = "[DEBUG] "
	case logrus.InfoLevel:
		levelLabel = "[INFO]  "
	case logrus.WarnLevel:
		colorPrefix = Yellow
		colorBoldPrefix = YellowBold
		colorSuffix = Reset
		levelLabel = "[WARN]  "
	case logrus.ErrorLevel:
		colorPrefix = Magenta
		colorBoldPrefix = MagentaBold
		colorSuffix = Reset
		levelLabel = "[ERROR] "
	case logrus.FatalLevel:
		colorPrefix = Red
		colorBoldPrefix = RedBold
		colorSuffix = Reset
		levelLabel = "[FATAL] "
	default:
		levelLabel = fmt.Sprintf("[%-7s]", level)
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// 获取调用者信息
	var callerLabel string = ""
	pc, file, line, ok := runtime.Caller(3) // 3 表示跳出logger包，取调用者信息
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		lastDot := strings.LastIndex(funcName, ".")
		if lastDot > 0 {
			if strings.Contains(funcName, ".func") { // 检查是否为匿名函数（func1、func2、func3等）
				mainName := funcName[:lastDot]
				lastDot = strings.LastIndex(mainName, ".")
				if lastDot <= 0 {
					lastDot = strings.LastIndex(mainName, "/")
				}
				if lastDot > 0 {
					funcName = mainName[lastDot+1:]
				} else {
					funcName = mainName
				}
			} else {
				funcName = funcName[lastDot+1:]
			}
		}

		shortFile := file
		if idx := len(file) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if file[i] == '/' {
					shortFile = file[i+1:]
					break
				}
			}
		}
		// if idx := len(file) - 1; idx >= 0 {
		// 	for i := 1; i <= idx; i++ {
		// 		if file[i] == '/' {
		// 			shortFile = file[i+1:]
		// 			break
		// 		}
		// 	}
		// }

		// msg = fmt.Sprintf("%s:%d(%s) %s", shortFile, line, funcName, msg)
		callerLabel = fmt.Sprintf("%s%s %s(%d)[%s]:", levelLabel, timestamp, shortFile, line, funcName)

	} else {
		callerLabel = fmt.Sprintf("%s%s:", levelLabel, timestamp)
	}

	if len(callerLabel)+len(msg) >= 145 {
		callerLabel = colorBoldPrefix + callerLabel + "\n" + colorSuffix
	} else {
		callerLabel = colorBoldPrefix + callerLabel + " " + colorSuffix
	}

	return callerLabel + colorPrefix + msg + colorSuffix
}

func updateLoggerFiles(date string) {
	// formatter := &logrus.TextFormatter{
	// 	TimestampFormat: "2006-01-02 15:04:05.000",
	// 	FullTimestamp:   true,
	// 	DisableQuote:    true, // 新增这一行
	// }
	formatter := &SimpleFormatter{}
	debugLogger.SetFormatter(formatter)
	debugLogger.SetLevel(logrus.DebugLevel)
	debugLogger.SetOutput(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/debug-%s.log", logDir, date),
		MaxAge:     maxAgeDays,
		MaxSize:    100,
		MaxBackups: 30,
		Compress:   true,
	})
	infoLogger.SetFormatter(formatter)
	infoLogger.SetLevel(logrus.InfoLevel)
	infoLogger.SetOutput(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/info-%s.log", logDir, date),
		MaxAge:     maxAgeDays,
		MaxSize:    100,
		MaxBackups: 30,
		Compress:   true,
	})
	warnLogger.SetFormatter(formatter)
	warnLogger.SetLevel(logrus.WarnLevel)
	warnLogger.SetOutput(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/warn-%s.log", logDir, date),
		MaxAge:     maxAgeDays,
		MaxSize:    100,
		MaxBackups: 30,
		Compress:   true,
	})
	errorLogger.SetFormatter(formatter)
	errorLogger.SetLevel(logrus.ErrorLevel)
	errorLogger.SetOutput(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/error-%s.log", logDir, date),
		MaxAge:     maxAgeDays,
		MaxSize:    100,
		MaxBackups: 30,
		Compress:   true,
	})
	fatalLogger.SetFormatter(formatter)
	fatalLogger.SetLevel(logrus.FatalLevel)
	fatalLogger.SetOutput(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/fatal-%s.log", logDir, date),
		MaxAge:     maxAgeDays,
		MaxSize:    100,
		MaxBackups: 30,
		Compress:   true,
	})
}

func checkDateAndUpdate() {
	date := time.Now().Format("2006-01-02")
	if lastDate.Load() == date {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	// 再次判断，防止并发重复切换
	if lastDate.Load() != date {
		lastDate.Store(date)
		updateLoggerFiles(date)
	}
}

func EnableDebugLog(enable bool) {
	debugEnabled = enable
}

func EnableWriteFileLog(enable bool) {
	writeFileEnabled = enable
}

func _debug(msg string) {
	if !debugEnabled {
		return
	}
	msg = _format(logrus.DebugLevel, msg)
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	debugLogger.Debug(msg)
}
func _info(msg string) {
	msg = _format(logrus.InfoLevel, msg)
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Info(msg)
	}
	infoLogger.Info(msg)
}
func _warn(msg string) {
	msg = _format(logrus.WarnLevel, msg)
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Warn(msg)
	}
	infoLogger.Warn(msg)
	warnLogger.Warn(msg)
}
func _error(msg string) {
	msg = _format(logrus.ErrorLevel, msg)
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Error(msg)
	}
	infoLogger.Error(msg)
	warnLogger.Error(msg)
	errorLogger.Error(msg)
}
func _fatal(msg string) {
	msg = _format(logrus.FatalLevel, fmt.Sprint(msg)+"\n"+string(debug.Stack()))
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Log(logrus.FatalLevel, msg)
	}
	infoLogger.Log(logrus.FatalLevel, msg)
	warnLogger.Log(logrus.FatalLevel, msg)
	errorLogger.Log(logrus.FatalLevel, msg)
	fatalLogger.Fatal(msg)
}

func Debug(format string, args ...interface{}) {
	if !debugEnabled {
		return
	}
	_debug(fmt.Sprintf(format, args...))
}

func Info(format string, args ...interface{}) {
	_info(fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	_warn(fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	_error(fmt.Sprintf(format, args...))
}

func Fatal(format string, args ...interface{}) {
	_fatal(fmt.Sprintf(format, args...))
}

type MyGormWriter struct{}

func (w *MyGormWriter) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// 根据 format 或 msg 内容判断级别，分发到 logger

	if strings.Contains(msg, "ERROR: operator does not exist:") { // postgres record not found
		// _writeRawToDebugFile(Blue + fmt.Sprintf("[Debug]  %s ", timestamp) + Reset + msg)
	} else if strings.Contains(msg, "[info]") {
		_writeRawToInfoFile(Green + fmt.Sprintf("[info]  %s ", timestamp) + Reset + msg)
	} else if strings.Contains(msg, "[warn]") {
		_writeRawToWarnFile(BlueBold + fmt.Sprintf("[warn]  %s ", timestamp) + Reset + msg)
	} else if strings.Contains(msg, "[error]") {
		_writeRawToErrorFile(Magenta + fmt.Sprintf("[error] %s ", timestamp) + Reset + msg)
	} else { // trace 级别
		if strings.Contains(msg, "record not found") { // confirmed
			// do nothing
		} else if strings.Contains(msg, "SLOW SQL") || // confirmed
			strings.Contains(msg, "failed to connect to") ||
			strings.Contains(msg, "connection refused") ||
			strings.Contains(msg, "broken pipe") ||
			strings.Contains(msg, "no such host") {
			_writeRawToWarnFile(YellowBold + fmt.Sprintf("[warn]  %s ", timestamp) + Reset + msg)
		} else if strings.Contains(msg, "panic") ||
			strings.Contains(msg, "Error") ||
			strings.Contains(msg, "context deadline exceeded") ||
			strings.Contains(msg, "context canceled") ||
			strings.Contains(msg, "duplicate entry") ||
			strings.Contains(msg, "timeout") ||
			strings.Contains(msg, "i/o timeout") {
			_writeRawToErrorFile(RedBold + fmt.Sprintf("[error] %s ", timestamp) + Reset + msg)
		} else { // 不确定的, 先往 error 文件里存, 但标签不加颜色
			_writeRawToErrorFile(fmt.Sprintf("[error] %s ", timestamp) + msg)
		}
	}
}

func _writeRawToDebugFile(msg string) {
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		fmt.Println(msg)
		debugLogger.Debug(msg)
	}
}

func _writeRawToInfoFile(msg string) {
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Info(msg)
	}
	infoLogger.Info(msg)
}

func _writeRawToWarnFile(msg string) {
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Warn(msg)
	}
	infoLogger.Warn(msg)
	warnLogger.Warn(msg)
}
func _writeRawToErrorFile(msg string) {
	fmt.Println(msg)
	if !writeFileEnabled {
		return
	}
	checkDateAndUpdate()
	if debugEnabled {
		debugLogger.Error(msg)
	}
	infoLogger.Error(msg)
	warnLogger.Error(msg)
	errorLogger.Error(msg)
}

/*
// 格式化输出（f）
func Debugf(format string, args ...interface{}) {
	if !DebugEnabled {
		return
	}
	_debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	_info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	_warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	_error(fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	_fatal(fmt.Sprintf(format, args...))
}

// 自动换行输出（ln）
func Debugln(args ...interface{}) {
	if !DebugEnabled {
		return
	}
	_debug(fmt.Sprintf("%v", args...))
}

func Infoln(args ...interface{}) {
	_info(fmt.Sprintf("%v", args...))
}

func Warnln(args ...interface{}) {
	_warn(fmt.Sprintf("%v", args...))
}

func Errorln(args ...interface{}) {
	_error(fmt.Sprintf("%v", args...))
}

func Fatalln(args ...interface{}) {
	_fatal(fmt.Sprintf("%v", args...))
}
*/

// 示例代码
// import "conso-wallet/pkg/logger"
// func main() {
// logger.InitLoggers("./logs", 7)
// logger.Info("服务启动")
// logger.Debug("调试信息")
// logger.Warn("警告信息")
// logger.Error("错误信息")
// logger.Fatal("致命错误")
// }
