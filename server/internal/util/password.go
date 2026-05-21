package util

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 用 bcrypt 加盐哈希，cost=10。
func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(h), err
}

// VerifyPassword 校验明文是否匹配 bcrypt 哈希。
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// RandPassword 生成长度为 n 的随机密码，字符集去除了易混淆字符（0/O/1/l/I）。
func RandPassword(n int) string {
	const charset = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}
