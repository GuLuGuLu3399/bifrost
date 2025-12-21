package http

import (
	"encoding/json"
	"net/http"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"google.golang.org/grpc/codes"
)

// ===========================================
// 错误响应结构 (遵循 Google API 错误规范)
// ===========================================

// ErrorResponse 标准错误响应
type ErrorResponse struct {
	Code    int           `json:"code"`              // gRPC 错误码
	Message string        `json:"message"`           // 人类可读的错误信息
	Details []ErrorDetail `json:"details,omitempty"` // 详细错误信息
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Type     string            `json:"@type"`              // 类型标识
	Reason   string            `json:"reason"`             // 错误原因 (机器可读)
	Domain   string            `json:"domain"`             // 错误域
	Metadata map[string]string `json:"metadata,omitempty"` // 附加元数据
}

// ===========================================
// 分页响应结构
// ===========================================

// PagedResponse 分页响应的通用字段
type PagedResponse struct {
	NextPageToken string `json:"next_page_token,omitempty"`
	TotalSize     int64  `json:"total_size,omitempty"`
}

// ===========================================
// 响应写入函数
// ===========================================

// JSON 写入 JSON 响应
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// 编码失败，记录日志，但 header 已发送
			return
		}
	}
}

// OK 返回成功响应 (HTTP 200)，直接返回业务数据，无信封包装
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created 返回创建成功响应 (HTTP 201)
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent 返回无内容响应 (HTTP 204)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ===========================================
// 错误响应函数
// ===========================================

// Error 返回错误响应，自动转换 xerr 到标准格式
func Error(w http.ResponseWriter, err error) {
	codeErr := xerr.FromError(err)

	grpcCode := mapBizCodeToGRPCCode(codeErr.GetCode())
	httpStatus := mapGRPCCodeToHTTPStatus(grpcCode)

	resp := ErrorResponse{
		Code:    int(grpcCode),
		Message: codeErr.GetMsg(),
	}

	JSON(w, httpStatus, resp)
}

// ErrorWithDetails 返回带详情的错误响应
func ErrorWithDetails(w http.ResponseWriter, err error, details ...ErrorDetail) {
	codeErr := xerr.FromError(err)

	grpcCode := mapBizCodeToGRPCCode(codeErr.GetCode())
	httpStatus := mapGRPCCodeToHTTPStatus(grpcCode)

	resp := ErrorResponse{
		Code:    int(grpcCode),
		Message: codeErr.GetMsg(),
		Details: details,
	}

	JSON(w, httpStatus, resp)
}

// NewErrorDetail 创建错误详情
func NewErrorDetail(reason, domain string, metadata map[string]string) ErrorDetail {
	return ErrorDetail{
		Type:     "type.googleapis.com/google.rpc.ErrorInfo",
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	}
}

// ===========================================
// 快捷错误响应
// ===========================================

// BadRequest 返回 400 错误
func BadRequest(w http.ResponseWriter, message string) {
	resp := ErrorResponse{
		Code:    int(codes.InvalidArgument),
		Message: message,
	}
	JSON(w, http.StatusBadRequest, resp)
}

// BadRequestWithDetails 返回 400 错误，带详情
func BadRequestWithDetails(w http.ResponseWriter, message string, details ...ErrorDetail) {
	resp := ErrorResponse{
		Code:    int(codes.InvalidArgument),
		Message: message,
		Details: details,
	}
	JSON(w, http.StatusBadRequest, resp)
}

// Unauthorized 返回 401 错误
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "身份认证失败"
	}
	resp := ErrorResponse{
		Code:    int(codes.Unauthenticated),
		Message: message,
	}
	JSON(w, http.StatusUnauthorized, resp)
}

// Forbidden 返回 403 错误
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "权限不足"
	}
	resp := ErrorResponse{
		Code:    int(codes.PermissionDenied),
		Message: message,
	}
	JSON(w, http.StatusForbidden, resp)
}

// NotFound 返回 404 错误
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "资源不存在"
	}
	resp := ErrorResponse{
		Code:    int(codes.NotFound),
		Message: message,
	}
	JSON(w, http.StatusNotFound, resp)
}

// Conflict 返回 409 错误
func Conflict(w http.ResponseWriter, message string, details ...ErrorDetail) {
	if message == "" {
		message = "资源冲突"
	}
	resp := ErrorResponse{
		Code:    int(codes.AlreadyExists),
		Message: message,
		Details: details,
	}
	JSON(w, http.StatusConflict, resp)
}

// TooManyRequests 返回 429 错误
func TooManyRequests(w http.ResponseWriter, message string) {
	if message == "" {
		message = "请求频率超限"
	}
	resp := ErrorResponse{
		Code:    int(codes.ResourceExhausted),
		Message: message,
	}
	JSON(w, http.StatusTooManyRequests, resp)
}

// InternalError 返回 500 错误
func InternalError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	resp := ErrorResponse{
		Code:    int(codes.Internal),
		Message: message,
	}
	JSON(w, http.StatusInternalServerError, resp)
}

// ===========================================
// 状态码映射
// ===========================================

// mapBizCodeToGRPCCode 将业务错误码映射到 gRPC 错误码
func mapBizCodeToGRPCCode(bizCode int) codes.Code {
	switch bizCode {
	case xerr.CodeOK:
		return codes.OK
	case xerr.CodeBadRequest, xerr.CodeValidation:
		return codes.InvalidArgument
	case xerr.CodeUnauthorized:
		return codes.Unauthenticated
	case xerr.CodeForbidden:
		return codes.PermissionDenied
	case xerr.CodeNotFound:
		return codes.NotFound
	case xerr.CodeConflict:
		return codes.AlreadyExists
	case xerr.CodeTimeout:
		return codes.DeadlineExceeded
	case xerr.CodeServiceUnavailable:
		return codes.Unavailable
	case xerr.CodeInternal:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// mapGRPCCodeToHTTPStatus 将 gRPC 错误码映射到 HTTP 状态码
func mapGRPCCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Internal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ===========================================
// 响应 Header 工具
// ===========================================

// SetTraceIDHeader 设置追踪 ID 响应头
func SetTraceIDHeader(w http.ResponseWriter, traceID string) {
	w.Header().Set("X-Trace-Id", traceID)
}

// SetRateLimitHeaders 设置限流相关响应头
func SetRateLimitHeaders(w http.ResponseWriter, limit, remaining int) {
	w.Header().Set("X-RateLimit-Limit", intToString(limit))
	w.Header().Set("X-RateLimit-Remaining", intToString(remaining))
}

// intToString 简单的整数转字符串
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	neg := n < 0
	if neg {
		n = -n
	}

	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}

	if neg {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}
