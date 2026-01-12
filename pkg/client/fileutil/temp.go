package fileutil

import "os"

// TempFile структура временного файла
type TempFile struct {
	*os.File
}

// NewTempFile создает новый временный файл
func NewTempFile(pattern string) (*TempFile, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}
	return &TempFile{File: file}, nil
}

// Close закрывает и удаляет временный файл
func (f *TempFile) Close() error {
	closeErr := f.File.Close()

	removeErr := os.Remove(f.File.Name())

	if closeErr != nil {
		return closeErr
	}
	return removeErr
}
