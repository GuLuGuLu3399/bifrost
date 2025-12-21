//此文件实现了一个基于Go标准库log包的日志记录器。

package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

type stdLogger struct {
	l     *log.Logger
	level Level
}

func NewStd() Logger {
	return &stdLogger{
		l:     log.New(os.Stdout, "", log.LstdFlags),
		level: InfoLevel,
	}
}

func NewStdWithLevel(level Level) Logger {
	return &stdLogger{
		l:     log.New(os.Stdout, "", log.LstdFlags),
		level: level,
	}
}

func (l *stdLogger) Debug(msg string, fields ...Field) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, fields)
	}
}

func (l *stdLogger) Info(msg string, fields ...Field) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, fields)
	}
}

func (l *stdLogger) Warn(msg string, fields ...Field) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, fields)
	}
}

func (l *stdLogger) Error(msg string, fields ...Field) {
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, fields)
	}
}

func (l *stdLogger) Fatal(msg string, fields ...Field) {
	l.log("FATAL", msg, fields)
	os.Exit(1)
}

// log 统一的日志打印实现
func (l *stdLogger) log(level, msg string, fields []Field) {
	if len(fields) == 0 {
		l.l.Printf("[%s] %s", level, msg)
		return
	}

	// 简化字段格式化
	var fieldStr string
	for _, f := range fields {
		switch f.Type {
		case fieldTypeString:
			fieldStr += fmt.Sprintf(" %s=%q", f.Key, f.Str)
		case fieldTypeInt:
			fieldStr += fmt.Sprintf(" %s=%d", f.Key, f.Int)
		case fieldTypeBool:
			fieldStr += fmt.Sprintf(" %s=%t", f.Key, f.Int == 1)
		case fieldTypeError:
			if err, ok := f.Interface.(error); ok {
				fieldStr += fmt.Sprintf(" %s=%q", f.Key, err.Error())
			} else {
				fieldStr += fmt.Sprintf(" %s=%q", f.Key, fmt.Sprintf("%v", f.Interface))
			}
		case fieldTypeAny:
			fieldStr += fmt.Sprintf(" %s=%q", f.Key, fmt.Sprintf("%v", f.Interface))
		default:
			fieldStr += fmt.Sprintf(" %s=%q", f.Key, fmt.Sprintf("%v", f.Interface))
		}
	}

	l.l.Printf("[%s] %s%s", level, msg, fieldStr)
}

func (l *stdLogger) With(...Field) Logger {
	return &stdLogger{
		l:     l.l,
		level: l.level,
	}
}

func (l *stdLogger) WithContext(context.Context) Logger {
	// 简化实现
	return l
}

func (l *stdLogger) SetLevel(level Level) {
	l.level = level
}

func (l *stdLogger) GetLevel() Level {
	return l.level
}

func (l *stdLogger) Sync() error {
	return nil
}

// NewStdWithFile 创建输出到文件的标准日志器
func NewStdWithFile(filename string, level Level) Logger {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// 如果无法打开文件，回退到标准输出
		return NewStdWithLevel(level)
	}

	return &stdLogger{
		l:     log.New(file, "", log.LstdFlags),
		level: level,
	}
}

// NewStdWithMultiOutput 创建多输出的标准日志器
func NewStdWithMultiOutput(outputs ...io.Writer) Logger {
	multiWriter := NewMultiWriter(outputs...)
	return &stdLogger{
		l:     log.New(multiWriter, "", log.LstdFlags),
		level: InfoLevel,
	}
}
