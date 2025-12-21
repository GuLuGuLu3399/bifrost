package xerr

import (
	"errors"
	"fmt"
	"sync"
)

// ==========================================
// 1. 构造函数
// ==========================================

// New 创建根因错误 (业务逻辑引发)
func New(code int, msg string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: msg,
		stack:   captureStack(),
	}
}

// Wrap 包装底层错误 (DB/IO 引发)
func Wrap(err error, code int, msg string) *CodeError {
	if err == nil {
		return nil
	}
	return &CodeError{
		Code:    code,
		Message: msg,
		cause:   err,
		stack:   captureStack(),
	}
}

// Wrapf 支持格式化的 Wrap
func Wrapf(err error, code int, format string, args ...interface{}) *CodeError {
	if err == nil {
		return nil
	}
	return &CodeError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		cause:   err,
		stack:   captureStack(),
	}
}

// ==========================================
// 2. 快捷构造器
// ==========================================

func BadRequest(format string, args ...interface{}) *CodeError {
	return New(CodeBadRequest, fmt.Sprintf(format, args...))
}

func Unauthorized(format string, args ...interface{}) *CodeError {
	return New(CodeUnauthorized, fmt.Sprintf(format, args...))
}

func Forbidden(format string, args ...interface{}) *CodeError {
	return New(CodeForbidden, fmt.Sprintf(format, args...))
}

func NotFound(format string, args ...interface{}) *CodeError {
	return New(CodeNotFound, fmt.Sprintf(format, args...))
}

func Internal(err error) *CodeError {
	// 内部错误必须保留堆栈/原错误，但对外展示 generic message
	return Wrap(err, CodeInternal, "internal server error")
}

// ==========================================
// 3. 智能转换 (FromError)
// ==========================================

type Mapper func(err error) *CodeError

var (
	mapperMu sync.RWMutex
	mappers  []Mapper
)

// RegisterMapper 注册自定义错误映射 (线程安全)
// 例如：将 gorm.ErrRecordNotFound 映射为 CodeNotFound
func RegisterMapper(m Mapper) {
	mapperMu.Lock()
	defer mapperMu.Unlock()
	mappers = append(mappers, m)
}

// FromError 将任意 error 转换为 *CodeError
func FromError(err error) *CodeError {
	if err == nil {
		return nil
	}

	// 1. 如果已经是 *CodeError，直接返回
	var codeErr *CodeError
	if As(err, &codeErr) {
		return codeErr
	}

	// 2. 尝试使用注册的 Mapper (如 GORM 错误)
	mapperMu.RLock()
	defer mapperMu.RUnlock()
	for i := len(mappers) - 1; i >= 0; i-- {
		if res := mappers[i](err); res != nil {
			return res
		}
	}

	// 3. 兜底：未知错误视为 Internal
	return Internal(err)
}

// As 是 xerr.As 的简化版本，用于检查错误类型
func As(err error, target interface{}) bool {
	// 这里简化实现，实际项目中可以使用标准库的 errors.As
	var codeErr *CodeError
	if errors.As(err, &codeErr) {
		if target, ok := target.(**CodeError); ok {
			*target = codeErr
			return true
		}
	}
	return false
}
