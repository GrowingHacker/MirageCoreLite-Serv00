package utils

import (
	"github.com/gin-contrib/sessions"
)

func IsSessionValid(session sessions.Session) bool {
	defer func() {
		_ = recover() // 捕获 panic，避免程序崩溃
	}()
	_ = session.Get("test") // 任意 key，用来触发 securecookie 解密
	return true
}
