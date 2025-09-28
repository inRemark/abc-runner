package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"abc-runner/servers/pkg/interfaces"
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel 解析日志级别字符串
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// Logger 日志记录器实现
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger(level string) *Logger {
	return &Logger{
		level:  ParseLogLevel(level),
		logger: log.New(os.Stdout, "", 0),
	}
}

// NewLoggerWithOutput 创建指定输出的日志记录器
func NewLoggerWithOutput(level string, output *os.File) *Logger {
	return &Logger{
		level:  ParseLogLevel(level),
		logger: log.New(output, "", 0),
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	if l.level <= DebugLevel {
		l.log(DebugLevel, msg, fields...)
	}
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	if l.level <= InfoLevel {
		l.log(InfoLevel, msg, fields...)
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	if l.level <= WarnLevel {
		l.log(WarnLevel, msg, fields...)
	}
}

// Error 记录错误日志
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	if l.level <= ErrorLevel {
		var mergedFields map[string]interface{}
		
		if len(fields) > 0 {
			mergedFields = make(map[string]interface{})
			for k, v := range fields[0] {
				mergedFields[k] = v
			}
		} else {
			mergedFields = make(map[string]interface{})
		}
		
		if err != nil {
			mergedFields["error"] = err.Error()
		}
		
		l.log(ErrorLevel, msg, mergedFields)
	}
}

// Fatal 记录致命错误日志
func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	var mergedFields map[string]interface{}
	
	if len(fields) > 0 {
		mergedFields = make(map[string]interface{})
		for k, v := range fields[0] {
			mergedFields[k] = v
		}
	} else {
		mergedFields = make(map[string]interface{})
	}
	
	if err != nil {
		mergedFields["error"] = err.Error()
	}
	
	l.log(FatalLevel, msg, mergedFields)
	os.Exit(1)
}

// log 内部日志记录方法
func (l *Logger) log(level LogLevel, msg string, fields ...map[string]interface{}) {
	logEntry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   msg,
		Fields:    make(map[string]interface{}),
	}
	
	// 合并字段
	if len(fields) > 0 {
		for k, v := range fields[0] {
			logEntry.Fields[k] = v
		}
	}
	
	// 格式化并输出日志
	output := l.formatLogEntry(logEntry)
	l.logger.Print(output)
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// formatLogEntry 格式化日志条目
func (l *Logger) formatLogEntry(entry LogEntry) string {
	// 使用JSON格式
	if len(entry.Fields) > 0 {
		jsonData, err := json.Marshal(entry)
		if err != nil {
			// 如果JSON序列化失败，使用简单格式
			return fmt.Sprintf("[%s] %s %s",
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.Level,
				entry.Message)
		}
		return string(jsonData)
	}
	
	// 简单格式
	return fmt.Sprintf("[%s] %s %s",
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		entry.Level,
		entry.Message)
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level string) {
	l.level = ParseLogLevel(level)
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// IsDebugEnabled 检查是否启用调试日志
func (l *Logger) IsDebugEnabled() bool {
	return l.level <= DebugLevel
}

// IsInfoEnabled 检查是否启用信息日志
func (l *Logger) IsInfoEnabled() bool {
	return l.level <= InfoLevel
}

// MultiLogger 多输出日志记录器
type MultiLogger struct {
	loggers []interfaces.Logger
}

// NewMultiLogger 创建多输出日志记录器
func NewMultiLogger(loggers ...interfaces.Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
	}
}

// AddLogger 添加日志记录器
func (ml *MultiLogger) AddLogger(logger interfaces.Logger) {
	ml.loggers = append(ml.loggers, logger)
}

// Debug 记录调试日志到所有记录器
func (ml *MultiLogger) Debug(msg string, fields ...map[string]interface{}) {
	for _, logger := range ml.loggers {
		logger.Debug(msg, fields...)
	}
}

// Info 记录信息日志到所有记录器
func (ml *MultiLogger) Info(msg string, fields ...map[string]interface{}) {
	for _, logger := range ml.loggers {
		logger.Info(msg, fields...)
	}
}

// Warn 记录警告日志到所有记录器
func (ml *MultiLogger) Warn(msg string, fields ...map[string]interface{}) {
	for _, logger := range ml.loggers {
		logger.Warn(msg, fields...)
	}
}

// Error 记录错误日志到所有记录器
func (ml *MultiLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	for _, logger := range ml.loggers {
		logger.Error(msg, err, fields...)
	}
}

// Fatal 记录致命错误日志到所有记录器
func (ml *MultiLogger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	for _, logger := range ml.loggers {
		// 注意：只有最后一个logger会调用os.Exit
		if logger == ml.loggers[len(ml.loggers)-1] {
			logger.Fatal(msg, err, fields...)
		} else {
			logger.Error(msg, err, fields...)
		}
	}
}

// FileLogger 文件日志记录器
type FileLogger struct {
	*Logger
	file *os.File
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(level, filePath string) (*FileLogger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	logger := NewLoggerWithOutput(level, file)
	
	return &FileLogger{
		Logger: logger,
		file:   file,
	}, nil
}

// Close 关闭文件日志记录器
func (fl *FileLogger) Close() error {
	if fl.file != nil {
		return fl.file.Close()
	}
	return nil
}

// RotatingLogger 轮转日志记录器
type RotatingLogger struct {
	*FileLogger
	maxSize   int64 // 最大文件大小（字节）
	maxFiles  int   // 最大文件数量
	filePath  string
	fileIndex int
}

// NewRotatingLogger 创建轮转日志记录器
func NewRotatingLogger(level, filePath string, maxSize int64, maxFiles int) (*RotatingLogger, error) {
	fileLogger, err := NewFileLogger(level, filePath)
	if err != nil {
		return nil, err
	}
	
	return &RotatingLogger{
		FileLogger: fileLogger,
		maxSize:    maxSize,
		maxFiles:   maxFiles,
		filePath:   filePath,
		fileIndex:  0,
	}, nil
}

// checkRotation 检查是否需要轮转
func (rl *RotatingLogger) checkRotation() error {
	if rl.file == nil {
		return nil
	}
	
	stat, err := rl.file.Stat()
	if err != nil {
		return err
	}
	
	if stat.Size() >= rl.maxSize {
		return rl.rotate()
	}
	
	return nil
}

// rotate 执行日志轮转
func (rl *RotatingLogger) rotate() error {
	// 关闭当前文件
	if err := rl.file.Close(); err != nil {
		return err
	}
	
	// 重命名当前文件
	rotatedPath := fmt.Sprintf("%s.%d", rl.filePath, rl.fileIndex)
	if err := os.Rename(rl.filePath, rotatedPath); err != nil {
		return err
	}
	
	// 创建新文件
	file, err := os.OpenFile(rl.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	
	rl.file = file
	rl.fileIndex++
	
	// 清理旧文件
	if rl.fileIndex >= rl.maxFiles {
		oldPath := fmt.Sprintf("%s.%d", rl.filePath, rl.fileIndex-rl.maxFiles)
		os.Remove(oldPath) // 忽略错误
	}
	
	return nil
}

// log 重写日志记录方法以支持轮转
func (rl *RotatingLogger) log(level LogLevel, msg string, fields ...map[string]interface{}) {
	// 检查是否需要轮转
	rl.checkRotation()
	
	// 调用父类方法
	rl.Logger.log(level, msg, fields...)
}