package authMiddleware

import (
	"mymodule/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		whiteList := map[string]bool{
			"/user/login":        true,
			"/static/login.html": true,
		}
		if whiteList[path] {
			ctx.Next()
			return
		}
		session := sessions.Default(ctx)
		if !utils.IsSessionValid(session) {
			// Session 解密失败或被篡改，重新初始化
			session.Set("init", true)
			_ = session.Save()
			ctx.Redirect(302, "/static/login.html")
			ctx.Abort()
			return
		}

		user := session.Get("user")
		if user == nil {
			ctx.Redirect(302, "/static/login.html")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
