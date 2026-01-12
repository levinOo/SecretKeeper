package fileutil

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// AllowedMimeTypes - белый список разрешенных типов.
// Лучше разрешать только то, что нужно, чем пытаться заблокировать всё плохое.
var AllowedMimeTypes = map[string]bool{
	"text/plain":               true, // .txt, .md, .env
	"application/json":         true, // .json
	"application/pdf":          true, // .pdf
	"image/jpeg":               true, // .jpg
	"image/png":                true, // .png
	"application/zip":          true, // .zip
	"application/octet-stream": true, // Общий бинарный тип (см. ниже примечание!)
}

var DangerousExtensions = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".sh": true, ".msi": true, ".com": true,
}

// DetectAndValidate проверяет файл на допустимые типы и безопасность
func DetectAndValidate(file *os.File) error {
	// Проверка расширения
	ext := strings.ToLower(filepath.Ext(file.Name()))
	if DangerousExtensions[ext] {
		return fmt.Errorf("оповещение безопасности: расширение файла '%s' не разрешено", ext)
	}

	// Читаем magic bytes
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {

	}

	// 3. Сбрасываем позицию чтения в начало файла!
	// Это критически важно, иначе io.ReadAll потом прочитает файл без начала.
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	mimeType := http.DetectContentType(buffer[:n])

	// Убираем параметры (например "text/plain; charset=utf-8" -> "text/plain")
	if i := strings.Index(mimeType, ";"); i != -1 {
		mimeType = mimeType[:i]
	}

	if !AllowedMimeTypes[mimeType] {
		return fmt.Errorf("тип файла '%s' не поддерживается", mimeType)
	}

	return nil
}

// CheckSize проверяет размер файла
func CheckSize(file *os.File) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() > 1024*1024*32 {
		return fmt.Errorf("размер файла слишком большой")
	}

	return nil
}

// GetFileSize возвращает размер файла
func GetFileSize(file *os.File) int64 {
	stat, err := file.Stat()
	if err != nil {
		return 0
	}

	return stat.Size()
}
