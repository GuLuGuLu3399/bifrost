package biz

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// 业务层错误定义
var (
	ErrPostNotFound              = xerr.New(xerr.CodeNotFound, "文章不存在")
	ErrSlugConflict              = xerr.New(xerr.CodeConflict, "文章标识已存在")
	ErrVersionConflict           = xerr.New(xerr.CodeConflict, "数据已被他人修改，请刷新重试")
	ErrUserNotFound              = xerr.New(xerr.CodeNotFound, "用户不存在")
	ErrUserAlreadyExists         = xerr.New(xerr.CodeConflict, "用户名或邮箱已存在")
	ErrInvalidCredentials        = xerr.New(xerr.CodeUnauthorized, "账号或密码错误")
	ErrPasswordMismatch          = xerr.New(xerr.CodeBadRequest, "旧密码错误")
	ErrUserInactive              = xerr.New(xerr.CodeForbidden, "账号已被禁用")
	ErrTagNotFound               = xerr.New(xerr.CodeNotFound, "标签不存在")
	ErrCategoryNotEmpty          = xerr.New(xerr.CodeValidation, "分类还关联文章，无法删除")
	ErrCategoryNotFound          = xerr.New(xerr.CodeNotFound, "分类不存在")
	ErrCommentNotFound           = xerr.New(xerr.CodeNotFound, "评论不存在")
	ErrCommentReplyForbidden     = xerr.New(xerr.CodeForbidden, "找不到父评论或无权限回复")
	ErrCommentConflict           = xerr.New(xerr.CodeConflict, "无权修改")
	ErrCommentParentPostMismatch = xerr.New(xerr.CodeValidation, "评论和父评论必须属于同一篇文章")
)

// Transaction 定义了事务的行为
// 业务层通过调用 ExecTx 来包裹需要原子执行的逻辑
type Transaction interface {
	ExecTx(ctx context.Context, fn func(ctx context.Context) error) error
}
