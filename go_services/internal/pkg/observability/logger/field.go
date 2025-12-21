//此文件定义了日志字段的构造器和相关类型

package logger

import (
	"fmt"
	"time"
)

// Field 定义了日志字段
type Field struct {
	Key       string
	Type      FieldType
	Int       int64
	Str       string
	Interface interface{}
}

type FieldType uint8

const (
	fieldTypeUnknown FieldType = iota
	fieldTypeString
	fieldTypeInt
	fieldTypeBool
	fieldTypeError
	fieldTypeAny
)

// -------------------------------------------------
// 构造器 (保持和 Zap API 一致，减少迁移成本)
// -------------------------------------------------

func String(key, val string) Field {
	return Field{Key: key, Type: fieldTypeString, Str: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Int8(key string, val int8) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Int16(key string, val int16) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: val}
}

func Uint(key string, val uint) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Uint8(key string, val uint8) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Uint16(key string, val uint16) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: int64(val)}
}

func Float32(key string, val float32) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

func Bool(key string, val bool) Field {
	var i int64
	if val {
		i = 1
	}
	return Field{Key: key, Type: fieldTypeBool, Int: i}
}

func Err(err error) Field {
	return Field{Key: "error", Type: fieldTypeError, Interface: err}
}

func Any(key string, val interface{}) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

// Duration 构造一个表示时间间隔的日志字段
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

// Time 构造一个表示时间的日志字段
func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

// Reflect 使用反射获取任意类型的值
func Reflect(key string, val interface{}) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

// Namespace 创建一个命名空间
func Namespace(key string) Field {
	return Field{Key: key, Type: fieldTypeString, Str: ""}
}

// Stack 构造一个堆栈字段
func Stack(key string) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: "stack"}
}

// -------------------------------------------------
// 高级构造器
// -------------------------------------------------

// Hex 将整数值表示为十六进制字符串
func Hex(key string, val uint64) Field {
	return Field{Key: key, Type: fieldTypeString, Str: "0x" + toHex(val)}
}

// toHex 将 uint64 转换为十六进制字符串
func toHex(val uint64) string {
	const hexDigits = "0123456789abcdef"
	if val == 0 {
		return "0"
	}

	var buf [16]byte
	i := 15
	for val > 0 {
		buf[i] = hexDigits[val%16]
		val /= 16
		i--
	}
	return string(buf[i+1:])
}

// ByteString 将字节数组表示为字符串
func ByteString(key string, val []byte) Field {
	return Field{Key: key, Type: fieldTypeAny, Interface: val}
}

// Stringer 实现了 fmt.Stringer 接口的类型
func Stringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Type: fieldTypeString, Str: val.String()}
}

// -------------------------------------------------
// 错误相关字段构造器
// -------------------------------------------------

// ErrorWithCode 创建带有错误码的错误字段
func ErrorWithCode(err error, code int) Field {
	return Field{
		Key:       "error",
		Type:      fieldTypeAny,
		Interface: map[string]interface{}{"message": err.Error(), "code": code},
	}
}

// ErrorStack 创建带有堆栈的错误字段
func ErrorStack(err error) Field {
	return Field{
		Key:       "error_stack",
		Type:      fieldTypeAny,
		Interface: err,
	}
}

// -------------------------------------------------
// 性能相关字段构造器
// -------------------------------------------------

// DurationMs 将时间转换为毫秒
func DurationMs(key string, val time.Duration) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: val.Milliseconds()}
}

// DurationNs 将时间转换为纳秒
func DurationNs(key string, val time.Duration) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: val.Nanoseconds()}
}

// Timestamp 创建时间戳字段
func Timestamp(key string, val time.Time) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: val.Unix()}
}

// TimestampMs 创建毫秒时间戳字段
func TimestampMs(key string, val time.Time) Field {
	return Field{Key: key, Type: fieldTypeInt, Int: val.UnixMilli()}
}

// -------------------------------------------------
// 系统相关字段构造器
// -------------------------------------------------

// PID 进程ID
func PID() Field {
	return Field{Key: "pid", Type: fieldTypeInt, Int: int64(getPID())}
}

// Hostname 主机名
func Hostname() Field {
	return Field{Key: "hostname", Type: fieldTypeString, Str: getHostname()}
}

// 获取进程ID的简化实现
func getPID() int {
	// 实际实现中可以使用 os.Getpid()
	return 0
}

// 获取主机名的简化实现
func getHostname() string {
	// 实际实现中可以使用 os.Hostname()
	return "unknown"
}
