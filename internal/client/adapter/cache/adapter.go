package cache

import (
	"os"
	"secretKeeper/internal/client/domain"
	pkgCache "secretKeeper/pkg/client/cache"
)

// FileCacheAdapter адаптирует pkg/client/cache.FileCache к usecase.CacheProvider.
type FileCacheAdapter struct {
	cache *pkgCache.FileCache
}

// NewFileCacheAdapter создает новый адаптер файлового кеша
func NewFileCacheAdapter(c *pkgCache.FileCache) *FileCacheAdapter {
	return &FileCacheAdapter{cache: c}
}

// Has проверяет наличие кеша
func (a *FileCacheAdapter) Has(id string) bool {
	return a.cache.Has(id)
}

// Get получает данные из кеша
func (a *FileCacheAdapter) Get(id string) (*os.File, *domain.SecretMeta, error) {
	file, meta, err := a.cache.Get(id)
	if err != nil {
		return nil, nil, err
	}

	return file, &domain.SecretMeta{
		ID:    id, // в метаданных кеша нет ID, но мы запрашивали по ID.
		Type:  domain.SecretType(meta.Type),
		Extra: meta.Extra,
	}, nil
}

// Put сохраняет данные в кеш
func (a *FileCacheAdapter) Put(id string, meta domain.SecretMeta) (*os.File, error) {
	pkgMeta := pkgCache.Metadata{
		Type:  string(meta.Type),
		Extra: meta.Extra,
	}
	return a.cache.Put(id, pkgMeta)
}
