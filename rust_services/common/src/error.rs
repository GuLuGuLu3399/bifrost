use std::fmt;

use http::StatusCode;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tonic::{Code, Status};

/// 业务错误码，保持与 Go 语义一致。
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[repr(u16)]
pub enum ErrorCode {
    Ok = 0,
    BadRequest = 400,
    Unauthorized = 401,
    Forbidden = 403,
    NotFound = 404,
    Conflict = 409,
    Validation = 422,
    Internal = 500,
    ServiceUnavailable = 503,
    Timeout = 504,
}

impl Default for ErrorCode {
    fn default() -> Self {
        ErrorCode::Internal
    }
}

impl ErrorCode {
    pub fn to_grpc(self) -> Code {
        match self {
            ErrorCode::Ok => Code::Ok,
            ErrorCode::BadRequest | ErrorCode::Validation => Code::InvalidArgument,
            ErrorCode::Unauthorized => Code::Unauthenticated,
            ErrorCode::Forbidden => Code::PermissionDenied,
            ErrorCode::NotFound => Code::NotFound,
            ErrorCode::Conflict => Code::AlreadyExists,
            ErrorCode::Timeout => Code::DeadlineExceeded,
            ErrorCode::ServiceUnavailable => Code::Unavailable,
            ErrorCode::Internal => Code::Internal,
        }
    }

    pub fn to_http(self) -> StatusCode {
        match self {
            ErrorCode::Ok => StatusCode::OK,
            ErrorCode::BadRequest | ErrorCode::Validation => StatusCode::BAD_REQUEST,
            ErrorCode::Unauthorized => StatusCode::UNAUTHORIZED,
            ErrorCode::Forbidden => StatusCode::FORBIDDEN,
            ErrorCode::NotFound => StatusCode::NOT_FOUND,
            ErrorCode::Conflict => StatusCode::CONFLICT,
            ErrorCode::Timeout => StatusCode::GATEWAY_TIMEOUT,
            ErrorCode::ServiceUnavailable => StatusCode::SERVICE_UNAVAILABLE,
            ErrorCode::Internal => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

/// 统一错误类型，兼容业务错误与内部错误。
#[derive(Debug, Error)]
#[error("{code:?}: {message}")]
pub struct CodeError {
    pub code: ErrorCode,
    pub message: String,
    #[source]
    pub source: Option<anyhow::Error>,
}

impl CodeError {
    pub fn new(code: ErrorCode, message: impl Into<String>) -> Self {
        Self { code, message: message.into(), source: None }
    }

    pub fn with_source(code: ErrorCode, message: impl Into<String>, source: impl Into<anyhow::Error>) -> Self {
        Self { code, message: message.into(), source: Some(source.into()) }
    }

    pub fn bad_request(msg: impl Into<String>) -> Self { Self::new(ErrorCode::BadRequest, msg) }
    pub fn unauthorized(msg: impl Into<String>) -> Self { Self::new(ErrorCode::Unauthorized, msg) }
    pub fn forbidden(msg: impl Into<String>) -> Self { Self::new(ErrorCode::Forbidden, msg) }
    pub fn not_found(msg: impl Into<String>) -> Self { Self::new(ErrorCode::NotFound, msg) }
    pub fn conflict(msg: impl Into<String>) -> Self { Self::new(ErrorCode::Conflict, msg) }
    pub fn validation(msg: impl Into<String>) -> Self { Self::new(ErrorCode::Validation, msg) }

    pub fn internal(msg: impl Into<String>) -> Self { Self::new(ErrorCode::Internal, msg) }
    pub fn internal_with(source: impl Into<anyhow::Error>, msg: impl Into<String>) -> Self {
        Self::with_source(ErrorCode::Internal, msg, source)
    }

    /// 将错误转为 tonic::Status，便于 gRPC 返回。
    pub fn to_status(&self) -> Status {
        Status::new(self.code.to_grpc(), self.message.clone())
    }

    /// 将错误转为 HTTP 错误响应。
    pub fn to_response(&self) -> ErrorResponse {
        ErrorResponse { code: self.code as u16, message: self.message.clone() }
    }
}

impl From<anyhow::Error> for CodeError {
    fn from(err: anyhow::Error) -> Self {
        CodeError { code: ErrorCode::Internal, message: "internal server error".to_string(), source: Some(err) }
    }
}

impl From<CodeError> for Status {
    fn from(err: CodeError) -> Self {
        err.to_status()
    }
}

/// HTTP/Gateway 统一错误信封，保持与 Go 侧 response.go 兼容字段。
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorResponse {
    pub code: u16,
    pub message: String,
}

impl fmt::Display for ErrorResponse {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {}", self.code, self.message)
    }
}

impl ErrorResponse {
    pub fn new(code: ErrorCode, message: impl Into<String>) -> Self {
        Self { code: code as u16, message: message.into() }
    }

    pub fn from_error(err: &CodeError) -> Self {
        err.to_response()
    }

    pub fn status_code(&self) -> StatusCode {
        StatusCode::from_u16(self.code).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR)
    }
}
