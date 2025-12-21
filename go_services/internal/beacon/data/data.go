package data

import (
	"github.com/gulugulu3399/bifrost/internal/pkg/cache"
	"github.com/gulugulu3399/bifrost/internal/pkg/database"
)

// Data 封装了 Beacon 共用的基础设施
type Data struct {
	db    *database.DB
	cache *cache.Client
}

func NewData(db *database.DB, cache *cache.Client) *Data {
	return &Data{
		db:    db,
		cache: cache,
	}
}

func (d *Data) DB() *database.DB {
	return d.db
}

func (d *Data) Cache() *cache.Client {
	return d.cache
}
