package proxyController

import (
	"mymodule/model"
	createservice "mymodule/service/proxyService/createService"
	deleteservice "mymodule/service/proxyService/deleteService"
	selectservice "mymodule/service/proxyService/selectService"
	startservice "mymodule/service/proxyService/startService"
	statusService "mymodule/service/proxyService/statusService"
	stopservice "mymodule/service/proxyService/stopService"
	updateservice "mymodule/service/proxyService/updateService"
	"mymodule/xraycoreHelper"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RigisterRouter(rg *gin.RouterGroup) {
	rg.GET("/start", start)
	rg.GET("/stop", stop)
	rg.GET("/create", create)
	rg.POST("/addConfig", addConfig)
	rg.GET("/select", selectAll)
	rg.GET("/checkStatus", checkStatus)
	rg.DELETE("/deleteConfig/:port", deleteConfig)
	rg.GET("/getByPort/:port", getByPort)
	// 新增更新路由
	rg.POST("/updateConfig/:oldPort", updateConfig)
}

var xray = &xraycoreHelper.XrayService{}

// 启动 xray
func start(c *gin.Context) {
	if ok, msg := startservice.Start(xray); !ok {
		c.JSON(500, gin.H{"success": false, "error": msg})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

// 停止 xray
func stop(c *gin.Context) {
	stopservice.Stop(xray)
	c.JSON(200, gin.H{"success": true})
}

// 跳转到新增页面
func create(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/create.html")
}

// 新增配置
func addConfig(c *gin.Context) {
	if ok, msg := createservice.Create(c.Request.Body); !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 查询所有
func selectAll(c *gin.Context) {
	data, err := selectservice.Select()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": data})
}

// 检查运行状态
func checkStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"running": statusService.CheckStatus(xray)})
}

// 删除配置
func deleteConfig(c *gin.Context) {
	portStr := c.Param("port")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "port 参数非法"})
		return
	}
	ok, msg := deleteservice.Delete(port)
	c.JSON(http.StatusOK, gin.H{"success": ok, "message": msg})
}

// 根据端口查询
func getByPort(c *gin.Context) {
	portStr := c.Param("port")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "port 参数非法"})
		return
	}
	list, err := selectservice.GetByPortWithCache(port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// 更新配置
func updateConfig(c *gin.Context) {
	// 1. 解析 URL 上的 oldPort
	oldPortStr := c.Param("oldPort")
	oldPort, err := strconv.Atoi(oldPortStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "oldPort 参数非法"})
		return
	}

	// 2. 绑定前端提交的 JSON 到 SelectList
	var newData model.SelectList
	if err := c.ShouldBindJSON(&newData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请求数据解析失败: " + err.Error()})
		return
	}

	// 3. 调用更新服务
	ok, msg := updateservice.Update(oldPort, newData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
