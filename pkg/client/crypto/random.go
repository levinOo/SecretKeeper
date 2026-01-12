package crypto

import (
	"crypto/rand"
	"io"
)

func GenerateRandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	// rand.Reader берет энтропию из /dev/urandom (OS), это безопасно.
	_, err := io.ReadFull(rand.Reader, b)
	return b, err
}
