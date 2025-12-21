package xerr

// 错误码定义
const (
	CodeOK                 = 0
	CodeBadRequest         = 400
	CodeUnauthorized       = 401
	CodeForbidden          = 403
	CodeNotFound           = 404
	CodeConflict           = 409
	CodeInternal           = 500
	CodeTimeout            = 504
	CodeValidation         = 422
	CodeServiceUnavailable = 503
)

// 错误码到字符串映射
var codeMessages = map[int]string{
	CodeOK:                 "OK",
	CodeBadRequest:         "Bad Request",
	CodeUnauthorized:       "Unauthorized",
	CodeForbidden:          "Forbidden",
	CodeNotFound:           "Not Found",
	CodeConflict:           "Conflict",
	CodeInternal:           "Internal Server Error",
	CodeTimeout:            "Gateway Timeout",
	CodeValidation:         "Validation Error",
	CodeServiceUnavailable: "Service Unavailable",
}

// StandardMessage 返回错误码的标准消息
func (e *CodeError) StandardMessage() string {
	if msg, exists := codeMessages[e.Code]; exists {
		return msg
	}
	return "Unknown Error"
}
