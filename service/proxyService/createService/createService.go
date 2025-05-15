// createservice/create.go
package createservice

import (
	"io"
	"log"
	"os"

	"mymodule/utils"
)

// Create 接收前端 POST 的 JSON，调用 FormatCfg 生成完整 Config，
// 加密并写入 AES_File
func Create(r io.Reader) (bool, string) {
	// 1. 读取请求体
	raw, err := io.ReadAll(r)
	if err != nil {
		log.Println("读取请求体失败：", err)
		return false, "读取请求体失败：" + err.Error()
	}

	// 2. 调用 FormatCfg 规范化：它会读旧配置、清理脏数据、合并或追加新条目并写回文件
	formattedPlain, err := utils.FormatCfg(raw)
	if err != nil {
		log.Println("FormatCfg 失败：", err)
		return false, "格式化配置失败：" + err.Error()
	}

	// 3. 对明文 JSON 再次加密（FormatCfg 已经写回了明文，这里加密后写入磁盘作最终存储）
	ciph, err := utils.Encrypt(formattedPlain)
	if err != nil {
		log.Println("加密失败：", err)
		return false, "加密失败：" + err.Error()
	}

	// 4. 确保配置目录存在
	if err := os.MkdirAll("config", 0755); err != nil {
		log.Println("创建目录失败：", err)
		return false, "创建目录失败：" + err.Error()
	}

	// 5. 写入加密后的配置文件
	if err := os.WriteFile(utils.AES_File, []byte(ciph), 0600); err != nil {
		log.Println("写入配置文件失败：", err)
		return false, "写入配置文件失败：" + err.Error()
	}

	return true, "保存成功"
}
