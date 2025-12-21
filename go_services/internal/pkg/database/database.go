package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/gulugulu3399/bifrost/internal/pkg/xerr"
)

// DB 封装 sqlx.DB，提供额外的事务支持
type DB struct {
	*sqlx.DB
}

// Close closes the underlying sql.DB connection pool.
// Call this on application shutdown.
func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

// New 创建新的数据库连接
func New(cfg *Config) (*DB, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	db, err := sqlx.Connect(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "failed to connect to database")
	}

	// 配置连接池
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, xerr.Wrap(err, xerr.CodeInternal, "failed to ping database")
	}

	return &DB{DB: db}, nil
}

// NewPostgres 创建 PostgreSQL 连接的便捷方法
func NewPostgres(host string, port int, user, password, dbname string, sslmode string) (*DB, error) {
	cfg := DefaultConfig()
	cfg.DSN = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	return New(cfg)
}

// RunInTx 在事务中执行函数，支持自动提交和回滚
func (db *DB) RunInTx(ctx context.Context, fn func(ctx context.Context, tx ExtContext) error) error {
	return db.RunInTxWithOptions(ctx, nil, fn)
}

// RunInTxWithOptions 使用指定选项在事务中执行函数
func (db *DB) RunInTxWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx ExtContext) error) error {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "failed to begin transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // 重新抛出 panic
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return xerr.Wrap(rbErr, xerr.CodeInternal, "failed to rollback transaction")
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return xerr.Wrap(err, xerr.CodeInternal, "failed to commit transaction")
	}

	return nil
}

// ExtContext 接口确保 Repository 方法既可以独立运行，也可以在事务中运行
// 这个接口聚合了 sqlx 的所有核心操作接口
type ExtContext interface {
	sqlx.ExtContext
	sqlx.ExecerContext
	sqlx.QueryerContext
	sqlx.PreparerContext
}

// Repository 基础 Repository 接口
// 所有 Repository 实现都应该使用 ExtContext 而不是 *sqlx.DB
// 这样可以在 Service 层灵活控制事务边界
type Repository interface {
	// 基础 CRUD 操作接口定义
}
