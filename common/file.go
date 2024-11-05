package common

import (
	"os"
	"path/filepath"
)

func fileExistsInDir(dir, filename string) (bool, error) {
	// 构造完整的文件路径
	filePath := filepath.Join(dir, filename)

	// 使用 os.Stat 获取文件信息
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// 判断是否是文件
	return !info.IsDir(), nil
}
