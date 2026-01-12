package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Metadata struct {
	Type  string            `json:"type"`
	Extra map[string]string `json:"extra"`
}

type FileCache struct {
	dir string
}

func NewFileCache(dir string) (*FileCache, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(userCacheDir, "secret-keeper", "blobs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return &FileCache{dir: dir}, nil
}

func (c *FileCache) Has(id string) bool {
	_, err := os.Stat(c.dataPath(id))
	return err == nil
}

func (c *FileCache) Get(id string) (*os.File, *Metadata, error) {
	metaBytes, err := os.ReadFile(c.metaPath(id))
	if err != nil {
		return nil, nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, nil, err
	}

	dataFile, err := os.Open(c.dataPath(id))
	if err != nil {
		return nil, nil, err
	}

	return dataFile, &meta, nil
}

func (c *FileCache) Put(id string, meta Metadata) (*os.File, error) {
	// Сохраняем мету
	metaBytes, _ := json.Marshal(meta)
	if err := os.WriteFile(c.metaPath(id), metaBytes, 0600); err != nil {
		return nil, err
	}

	// Открываем файл для записи данных (если был - перезапишем)
	return os.Create(c.dataPath(id))
}

func (c *FileCache) dataPath(id string) string { return filepath.Join(c.dir, id+".bin") }
func (c *FileCache) metaPath(id string) string { return filepath.Join(c.dir, id+".meta") }
