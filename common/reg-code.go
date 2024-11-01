package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// generateRegCode 生成一个基于SHA-256的注册码，长度为16字节（32个十六进制字符）
func GenerateRegCode() (string, error) {
	// 生成随机数据用于哈希输入
	randomData := make([]byte, 32) // 通过增加byte数量增大前端的随机性
	if _, err := rand.Read(randomData); err != nil {
		return "", err
	}

	// 使用SHA-256进行哈希
	hasher := sha256.New()
	hasher.Write(randomData)
	fullHash := hasher.Sum(nil)

	// 将哈希值转换成十六进制字符串
	fullHashHex := hex.EncodeToString(fullHash)

	// 截取部分哈希作为注册码
	return fullHashHex[:32], nil // 32个字符，16个字节
}
