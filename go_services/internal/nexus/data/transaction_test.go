package data

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gulugulu3399/bifrost/internal/pkg/database"
	"github.com/jmoiron/sqlx"
)

func newSQLMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	// Close returns error; tests don't care, but we must not ignore it.
	t.Cleanup(func() { _ = db.Close() })

	return sqlx.NewDb(db, "sqlmock"), mock
}

func TestTransaction_ExecTx_RollbackOnPanic(t *testing.T) {
	sqlxDB, mock := newSQLMock(t)
	db := &database.DB{DB: sqlxDB}
	tr := &transaction{db: db}

	mock.ExpectBegin()
	mock.ExpectRollback()

	defer func() {
		p := recover()
		if p == nil {
			t.Fatalf("expected panic")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations: %v", err)
		}
	}()

	_ = tr.ExecTx(context.Background(), func(ctx context.Context) error {
		panic("boom")
	})
}

func TestTransaction_ExecTx_ReturnsCommitError(t *testing.T) {
	sqlxDB, mock := newSQLMock(t)
	db := &database.DB{DB: sqlxDB}
	tr := &transaction{db: db}

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(assertErr{})

	err := tr.ExecTx(context.Background(), func(ctx context.Context) error { return nil })
	if err == nil {
		t.Fatalf("expected error")
	}
	var ae assertErr
	if !errors.As(err, &ae) {
		t.Fatalf("expected assertErr, got %T", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "assert" }
