package data

import (
	"context"

	"github.com/gulugulu3399/bifrost/internal/nexus/biz"
	"github.com/gulugulu3399/bifrost/internal/pkg/database"
)

// contextTxKey 用于在 Context 中存储事务对象
type contextTxKey struct{}

type transaction struct {
	db *database.DB
}

func NewTransaction(db *database.DB) biz.Transaction {
	return &transaction{db: db}
}

// ExecTx 实现事务逻辑
func (t *transaction) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.RunInTx(ctx, func(ctx context.Context, tx database.ExtContext) error {
		// 将 tx 注入 context，供 Data.DB(ctx) 取用
		ctxWithTx := context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctxWithTx)
	})
}
