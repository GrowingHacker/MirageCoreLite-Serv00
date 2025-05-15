package utils

import (
	"log"
	"os"
)

func checkPath(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err // 路径不存在
		}
		// 其他错误（如权限问题）
		log.Println("发生错误:", err)
		return nil, err
	}
	return info, nil
}

func DirExists(path string) (bool, error) {
	info, err := checkPath(path)
	if err != nil {
		return false, nil
	}
	return info.IsDir(), nil // 存在且是目录
}

func FileExists(path string) (bool, error) {
	info, err := checkPath(path)
	if err != nil {
		return false, nil
	}
	return !info.IsDir(), nil // 存在且不是目录（即是文件）
}
