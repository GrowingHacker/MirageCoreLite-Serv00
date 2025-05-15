package utils

import (
	"encoding/json"
	"os"
)

// 修改配置文件中的端口
func PatchXrayPort(inputPath, outputPath, port string) error {
	// 使用 os.ReadFile 替代 ioutil.ReadFile
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	// 反序列化配置为 map
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	// 遍历 inbounds，替换端口
	if inbounds, ok := config["inbounds"].([]interface{}); ok {
		for _, inbound := range inbounds {
			if ib, ok := inbound.(map[string]interface{}); ok {
				ib["port"] = port // 修改端口
			}
		}
	}

	// 写入修改后的配置
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	// 使用 os.WriteFile 替代 ioutil.WriteFile
	return os.WriteFile(outputPath, newData, 0644)
}
