package main

import (
	"crypto/rand"
	"embed"
	"io/fs"
	"log"
	authMiddleware "mymodule/middleware"
	"mymodule/router"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	r := gin.Default()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	store := cookie.NewStore(key)
	store.Options(sessions.Options{
		MaxAge:   3600 * 5, //一次Session有效期为5小时
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	})
	r.Use(sessions.Sessions("userSession", store))
	r.Use(authMiddleware.AuthRequired()) //全局拦截路由请求，验证session

	// 把 embed 的 FS 转成以 static 为根的子 FS
	staticFS, err := fs.Sub(staticFiles, "static")
	fs := http.FS(staticFS)
	r.StaticFS("/static", fs)
	if err != nil {
		log.Fatal("无法加载 embedded static:", err)
	}

	router.SetUpRouter(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "16558"
	}
	r.Run(":" + port)

}
