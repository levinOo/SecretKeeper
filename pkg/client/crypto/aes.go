package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// CryptoManager реализует интерфейс CryptoService.
// Он хранит ключ сессии, чтобы не передавать его в каждый метод.
type CryptoManager struct {
	key []byte // Master Key (32 байта для AES-256)
}

// NewCryptoManager создает новый менеджер с уже вычисленным ключом.
func NewCryptoManager(key []byte) *CryptoManager {
	if len(key) != 32 {
		panic("crypto: key length must be 32 bytes")
	}
	return &CryptoManager{key: key}
}

// Encrypt реализует шифрование AES-GCM для небольших данных (в памяти)
func (m *CryptoManager) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal добавляет nonce в начало зашифрованных данных
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt расшифровывает данные
func (m *CryptoManager) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]

	return gcm.Open(nil, nonce, actualCiphertext, nil)
}

// EncryptStream реализует потоковое шифрование (AES-OFB) для файлов
func (m *CryptoManager) EncryptStream(in io.Reader, out io.Writer) error {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return err
	}

	// Генерируем IV (Vector инициализации) для потока
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Пишем IV в начало файла (он нужен для расшифровки)
	if _, err := out.Write(iv); err != nil {
		return err
	}

	stream := cipher.NewOFB(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: out}

	// Копируем данные через шифрующий поток
	if _, err := io.Copy(writer, in); err != nil {
		return err
	}

	return nil
}
