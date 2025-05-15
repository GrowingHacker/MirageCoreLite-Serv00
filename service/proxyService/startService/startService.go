package startservice

import (
	"mymodule/utils"
	"mymodule/xraycoreHelper"
	"os"
)

func Start(xray *xraycoreHelper.XrayService) (bool, string) {
	encData, err := os.ReadFile(utils.AES_File)
	if err != nil {
		return false, "配置文件不存在"
	}
	jsonData, err := utils.Decrypt(string(encData))
	if err != nil {
		return false, "解密失败"
	}
	//fmt.Println("解密后：\n", string(jsonData))
	if err := xray.Start(string(jsonData)); err != nil {
		return false, "启动失败: " + err.Error()
	}
	return true, "启动成功"
}
