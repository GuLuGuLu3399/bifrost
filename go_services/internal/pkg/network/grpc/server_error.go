package grpc

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorInterceptor 统一错误处理拦截器
func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)

		if err != nil {
			// 将业务错误 (xerr) 转换为 gRPC 状态码
			return resp, wrapServerError(err)
		}

		return resp, nil
	}
}

// wrapServerError 处理业务错误到 gRPC 状态码的映射
func wrapServerError(err error) error {
	// 尝试解析自定义的 xerr
	bizErr := xerr.FromError(err)
	if bizErr != nil {
		// 根据业务 Code 映射 gRPC 状态码
		grpcCode := mapBizCodeToGRPCCode(bizErr.GetCode())
		st := status.New(grpcCode, bizErr.GetMsg())
		return st.Err()
	}

	return status.Error(codes.Unknown, err.Error())
}

// mapBizCodeToGRPCCode 将业务错误码映射到 gRPC 错误码
// 遵循 Bifrost 接口规范
func mapBizCodeToGRPCCode(bizCode int) codes.Code {
	switch bizCode {
	case xerr.CodeOK:
		return codes.OK
	case xerr.CodeBadRequest, xerr.CodeValidation:
		return codes.InvalidArgument // HTTP 400
	case xerr.CodeUnauthorized:
		return codes.Unauthenticated // HTTP 401
	case xerr.CodeForbidden:
		return codes.PermissionDenied // HTTP 403
	case xerr.CodeNotFound:
		return codes.NotFound // HTTP 404
	case xerr.CodeConflict:
		return codes.AlreadyExists // HTTP 409
	case xerr.CodeTimeout:
		return codes.DeadlineExceeded // HTTP 504
	case xerr.CodeServiceUnavailable:
		return codes.Unavailable // HTTP 503
	case xerr.CodeInternal:
		return codes.Internal // HTTP 500
	default:
		return codes.Unknown
	}
}
