package logger

import (
	"context"
	"errors"
	"sync"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// Level 日志级别
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l Level) String() string {
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

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger // 明确使用 contextx.Context
	SetLevel(level Level)
	GetLevel() Level
	Sync() error
}

// Config 日志配置 (已移除轮转相关的废弃字段)
type Config struct {
	Level       Level  `yaml:"level"`
	Format      string `yaml:"format"`   // json 或 console
	Output      string `yaml:"output"`   // stdout, stderr, file
	Filename    string `yaml:"filename"` // 如果 Output 是 file，则需要此项
	CallerSkip  int    `yaml:"caller_skip"`
	EnableColor bool   `yaml:"enable_color"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:       InfoLevel,
		Format:      "json",
		Output:      "stdout",
		CallerSkip:  1,
		EnableColor: true,
	}
}

// Global 实例管理
var (
	globalMu sync.RWMutex
	global   = NewStd()
)

func SetGlobal(l Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = l
}

func Global() Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return global
}

// -------------------------------------------------
// 快捷函数 (直接代理给 Global)
// -------------------------------------------------

func Debug(msg string, fields ...Field) { Global().Debug(msg, fields...) }
func Info(msg string, fields ...Field)  { Global().Info(msg, fields...) }
func Warn(msg string, fields ...Field)  { Global().Warn(msg, fields...) }
func Error(msg string, fields ...Field) { Global().Error(msg, fields...) }
func Fatal(msg string, fields ...Field) { Global().Fatal(msg, fields...) }
func Sync() error                       { return Global().Sync() }

func With(fields ...Field) Logger {
	return Global().With(fields...)
}

func WithContext(ctx context.Context) Logger {
	return Global().WithContext(ctx)
}

func SetLevel(level Level) { Global().SetLevel(level) }
func GetLevel() Level      { return Global().GetLevel() }

// -------------------------------------------------
// 错误日志辅助函数
// -------------------------------------------------

// ErrorWithStack 记录带有堆栈的错误
func ErrorWithStack(err error, msg string, fields ...Field) {
	if err == nil {
		return
	}

	// 针对 xerr.CodeError 的特殊处理
	var codeErr *xerr.CodeError
	if errors.As(err, &codeErr) {
		// 预分配 slice 容量，避免多次扩容
		stackFields := make([]Field, 0, len(fields)+3)
		stackFields = append(stackFields, fields...)
		stackFields = append(stackFields,
			Int("error_code", codeErr.Code),
			String("error_message", codeErr.Message),
		)

		if stack := codeErr.StackTrace(); len(stack) > 0 {
			stackFields = append(stackFields, Any("stack_trace", stack))
		}

		Global().Error(msg, stackFields...)
		return
	}

	// 普通错误
	Global().Error(msg, append(fields, String("error", err.Error()))...)
}

// WarnWithError 记录警告级别的错误
func WarnWithError(err error, msg string, fields ...Field) {
	if err == nil {
		return
	}
	Global().Warn(msg, append(fields, String("error", err.Error()))...)
}
