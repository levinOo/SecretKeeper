package crypto

import "golang.org/x/crypto/argon2"

// Генерация ключа на основе пароля и соли
func DeriveKey(pass string, salt []byte) []byte {
	// Параметры Argon2id (стандарт OWASP 2024):
	// Time (1): Количество проходов. 1 достаточно для Argon2id.
	// Memory (64*1024): 64 МБ памяти. Это делает перебор на GPU очень дорогим.
	// Threads (4): Используем 4 потока процессора.
	// KeyLen (32): Нам нужно ровно 32 байта для AES-256.
	return argon2.IDKey([]byte(pass), salt, 1, 64*1024, 4, 32)
}
