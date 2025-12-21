package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/gulugulu3399/bifrost/internal/pkg/database"
	"github.com/redis/go-redis/v9"
)

// Data 数据层依赖容器
type Data struct {
	db    *database.DB
	redis *redis.Client
}

// NewData 创建数据层依赖
func NewData(db *database.DB, redis *redis.Client) *Data {
	return &Data{
		db:    db,
		redis: redis,
	}
}

// DB 获取当前上下文的执行器 (可能是 *sqlx.DB 也可能是 *sqlx.Tx)
func (d *Data) DB(ctx context.Context) database.ExtContext {
	if d == nil {
		return nil
	}
	if tx, ok := ctx.Value(contextTxKey{}).(database.ExtContext); ok && tx != nil {
		return tx
	}
	return d.db
}

// RawDB 返回底层连接池（仅限极少数必须使用 *database.DB 的场景）
func (d *Data) RawDB() *database.DB {
	return d.db
}

// Redis 获取 Redis 客户端
func (d *Data) Redis() *redis.Client {
	return d.redis
}

// Close 关闭数据库连接
func (d *Data) Close() error {
	if d == nil {
		return nil
	}
	if d.redis != nil {
		_ = d.redis.Close()
	}
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// ===========================================
// 辅助函数
// ===========================================

// ---------------------------
// Null -> Value (读取方向)
// ---------------------------

// 辅助：把 sql.NullString 转为普通 string
func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// 辅助：把 sql.NullInt64 转为 int64
func nullInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}

// 辅助：把 sql.NullTime 转为 *time.Time
func nullTime(at sql.NullTime) *time.Time {
	if at.Valid {
		return &at.Time
	}
	return nil
}

// ---------------------------
// Value -> Null (写入方向)
// ---------------------------

// stringToNullString 将非空字符串转换为 sql.NullString（空字符串 => NULL）
func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// int64ToNullInt64 将非 0 int64 转换为 sql.NullInt64（0 => NULL）
func int64ToNullInt64(v int64) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}

// ptrTimeToNullTime 将 *time.Time 转换为 sql.NullTime (nil => NULL)
func ptrTimeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// 辅助方法：Enum 转 DB String (适配你的 SQL VARCHAR 定义)
// NOTE: 该项目 posts.status 在数据层使用 int32 存储（见 postPO.Status），不需要 string 映射。
// 如未来某张表确实使用 VARCHAR enum，请在对应 repo 文件中实现局部映射，避免 Data 层出现业务 repo 方法。
