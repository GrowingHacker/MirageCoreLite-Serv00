package userController

import (
	"fmt"
	"mymodule/service/userService/loginService"
	usersettingservice "mymodule/service/userService/userSettingService"
	"mymodule/utils"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type updateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}
type passwordRequest struct {
	Password string `json:"password"` // 要与前端字段名匹配
}

func RigisterRouter(rg *gin.RouterGroup) {
	rg.POST("/login", login)
	rg.POST("/getToken", getToken)
	rg.PUT("/update", updateUserConfig)
}
func login(con *gin.Context) {
	var request loginRequest
	if err := con.ShouldBindJSON(&request); err != nil {
		con.JSON(http.StatusBadRequest, gin.H{"erro": fmt.Sprintf("读取请求失败：%v", err)})
	}

	tag, err := loginService.Login(request.Username, request.Password)
	if tag {
		session := sessions.Default(con)
		//校验session是否有效，并重置无效的session
		if !utils.IsSessionValid(session) {
			session.Set("init", true)
		}

		session.Set("user", request.Username)
		session.Save()
		con.JSON(http.StatusOK, gin.H{
			"code":    200,
			"success": true,
			"message": "登录成功！",
			"data": gin.H{
				"username": request.Username,
			},
		})
		return
	}

	con.JSON(http.StatusUnauthorized, gin.H{"erro": "登录失败！" + err})

}

func getToken(con *gin.Context) {
	var request passwordRequest
	if err := con.ShouldBindJSON(&request); err != nil {
		con.JSON(http.StatusBadRequest, gin.H{"erro": fmt.Sprintf("读取请求失败：%v", err)})
	}
	session := sessions.Default(con)
	user := session.Get("user")
	if userStr, ok := user.(string); ok {
		success, token, err := usersettingservice.VerifyOldPwd(request.Password, userStr)
		if success {
			con.JSON(http.StatusOK, gin.H{"token": token, "erro": err})
		} else {
			con.JSON(http.StatusOK, gin.H{"token": token})
		}
	} else {
		con.JSON(http.StatusOK, gin.H{"erro": "session参数格式错误"})
	}

}

func updateUserConfig(con *gin.Context) {
	var request updateRequest
	session := sessions.Default(con)
	user := session.Get("user")
	if userStr, ok := user.(string); ok {
		if err := con.ShouldBindJSON(&request); err != nil {
			con.JSON(http.StatusBadRequest, gin.H{"erro": fmt.Sprintf("读取请求失败：%v", err)})
		}
		success, err := usersettingservice.ReSet(request.Username, request.Password, userStr, request.Token)
		//重置会话
		if success {
			session.Set("init", true)
		}
		con.JSON(http.StatusOK, gin.H{"success": success, "erro": err})
	} else {
		con.JSON(http.StatusOK, gin.H{"erro": "session参数格式错误"})
	}

}
