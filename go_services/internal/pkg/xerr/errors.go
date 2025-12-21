package xerr

import (
	"fmt"
	"runtime"
)

// CodeError 统一业务错误
type CodeError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`

	// cause 持有原始错误 (如 db error)，仅用于日志记录，不暴露给前端
	cause error

	// stack 持有调用堆栈
	stack []uintptr
}

// Error 实现标准 error 接口
// 格式设计为对日志友好：[错误码] 业务消息: 底层原因
// 使用指针接收器 *CodeError 以避免复制并保持一致性
func (e *CodeError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持 Go 1.13+ 错误链 (xerr.Is/As)
func (e *CodeError) Unwrap() error {
	return e.cause
}

// GetMsg 返回错误消息
func (e *CodeError) GetMsg() string {
	return e.Message
}

// GetCode 返回错误码
func (e *CodeError) GetCode() int {
	return e.Code
}

// StackTrace 返回堆栈信息
func (e *CodeError) StackTrace() []string {
	if len(e.stack) == 0 {
		return nil
	}

	frames := runtime.CallersFrames(e.stack)
	var stack []string
	for {
		frame, more := frames.Next()
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return stack
}

// captureStack 捕获当前调用堆栈
func captureStack() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	// skip 3: runtime.Callers -> captureStack -> New/Wrap -> Caller
	n := runtime.Callers(3, pcs[:])
	return pcs[0:n]
}
