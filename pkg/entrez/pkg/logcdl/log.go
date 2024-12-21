package logcdl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	TIP
	WARN
	ERROR
	PANIC
	FATAL
	SUCCESS
)

// ANSI 颜色代码
const (
	colorRed     = "\033[31m"   // 红色
	colorBoldRed = "\033[1;31m" // 高亮红色
	colorGreen   = "\033[32m"   // 绿色
	colorYellow  = "\033[33m"   // 黄色
	colorBlue    = "\033[34m"   // 蓝色
	colorCyan    = "\033[36m"   // 青色
	colorGray    = "\033[90m"   // 灰色
	colorReset   = "\033[0m"    // 重置
)

// LevelInfo 日志级别信息
type LevelInfo struct {
	prefix string
	color  string
}

// getLevelInfo 获取日志级别信息
func getLevelInfo(level Level) LevelInfo {
	switch level {
	case DEBUG:
		return LevelInfo{prefix: "[DEBUG]", color: colorCyan}
	case INFO:
		return LevelInfo{prefix: "[INFO]", color: ""}
	case TIP:
		return LevelInfo{prefix: "[TIP]", color: colorBlue}
	case WARN:
		return LevelInfo{prefix: "[WARN]", color: colorYellow}
	case ERROR:
		return LevelInfo{prefix: "[ERROR]", color: colorRed}
	case PANIC:
		return LevelInfo{prefix: "[PANIC]", color: colorBoldRed}
	case FATAL:
		return LevelInfo{prefix: "[FATAL]", color: colorBoldRed}
	case SUCCESS:
		return LevelInfo{prefix: "[SUCCESS]", color: colorGreen}
	default:
		return LevelInfo{prefix: "[INFO]", color: ""}
	}
}

// Logger 封装了日志记录器
type Logger struct {
	console  *log.Logger
	file     *log.Logger
	logFile  *os.File
	minLevel Level   // 最小日志级别
	options  Options // 保存配置选项
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Options 日志选项
type Options struct {
	MinLevel    Level  // 最小日志级别
	LogDir      string // 日志目录
	LogFileName string // 日志文件名格式
	TimeFormat  string // 时间格式
}

// DefaultOptions 默认选项
var DefaultOptions = Options{
	MinLevel:    DEBUG,
	LogDir:      "logs",
	LogFileName: "clinvardl_%s.log",
	TimeFormat:  "2006-01-02",
}

// NewLogger 创建新的日志记录器
func NewLogger(opts ...Options) (*Logger, error) {
	options := DefaultOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	logger := &Logger{
		console:  log.New(os.Stdout, "", log.LstdFlags),
		minLevel: options.MinLevel,
		options:  options,
	}

	// 创建日志文件
	if err := logger.createLogFile(); err != nil {
		return nil, err
	}

	return logger, nil
}

// InitLogger 初始化日志记录器
func InitLogger(opts ...Options) error {
	var err error
	once.Do(func() {
		var logger *Logger
		logger, err = NewLogger(opts...)
		if err != nil {
			return
		}

		defaultLogger = logger
		// 将标准日志输出重定向到文件
		log.SetOutput(logger.logFile)
		log.SetFlags(log.LstdFlags)
	})

	return err
}

// Close 关闭日志文件
func Close() error {
	if defaultLogger != nil && defaultLogger.logFile != nil {
		return defaultLogger.logFile.Close()
	}
	return nil
}

// log 内部日志方法
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	info := getLevelInfo(level)
	msg := fmt.Sprintf(format, args...)

	// 控制台输出，只有在有颜色时才添加颜色代码
	if info.color != "" {
		l.console.Printf("%s%s %s%s", info.color, info.prefix, msg, colorReset)
	} else {
		l.console.Printf("%s %s", info.prefix, msg)
	}

	// 文件普通输出
	l.file.Printf("%s %s", info.prefix, msg)

	// 添加对 PANIC 和 FATAL 的处理
	if level == FATAL {
		l.logFile.Close()
		os.Exit(1)
	} else if level == PANIC {
		l.logFile.Close()
		panic(msg)
	}
}

// createLogDir 创建日志目录
func (l *Logger) createLogDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	logDir := filepath.Join(currentDir, l.options.LogDir)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create log directory: %v", err)
	}

	return logDir, nil
}

// createLogFile 创建日志文件
func (l *Logger) createLogFile() error {
	// 创建日志目录
	logDir, err := l.createLogDir()
	if err != nil {
		return err
	}

	// 创建日志文件
	timestamp := time.Now().Format(l.options.TimeFormat)
	logFileName := fmt.Sprintf(l.options.LogFileName, timestamp)
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	l.logFile = logFile
	l.file = log.New(logFile, "", log.LstdFlags)
	return nil
}

// Logger 的方法
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Tip(format string, args ...interface{}) {
	l.log(TIP, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Panic(format string, args ...interface{}) {
	l.log(PANIC, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

func (l *Logger) Success(format string, args ...interface{}) {
	l.log(SUCCESS, format, args...)
}

// 包级别的方法只是对默认 logger 的简单封装
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Tip(format string, args ...interface{}) {
	defaultLogger.Tip(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Panic(format string, args ...interface{}) {
	defaultLogger.Panic(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

func Success(format string, args ...interface{}) {
	defaultLogger.Success(format, args...)
}
