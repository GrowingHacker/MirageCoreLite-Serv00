package router

import (
	"mymodule/controller/proxyController"
	"mymodule/controller/userController"

	"github.com/gin-gonic/gin"
)

func SetUpRouter(r *gin.Engine) {
	userController.RigisterRouter(r.Group("/user")) //注册用户路由,并且设置路由组为user（即前缀为/user）
	proxyController.RigisterRouter(r.Group("/proxy"))
}
