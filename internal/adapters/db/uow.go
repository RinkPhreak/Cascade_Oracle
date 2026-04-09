package db

import (
	"context"

	"gorm.io/gorm"

	"cascade/internal/application/port"
)

type txKey struct{}

type gormUnitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new SQL transaction wrapper.
func NewUnitOfWork(db *gorm.DB) port.UnitOfWork {
	return &gormUnitOfWork{db: db}
}

// Execute wraps logic inside a thread-safe context transaction.
func (u *gormUnitOfWork) Execute(ctx context.Context, fn func(context.Context) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// ExtractDB seamlessly resolves the injected transaction or falls back to standard GORM connection.
func ExtractDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
