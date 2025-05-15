package updateservice

import (
	"encoding/json"
	"fmt"
	"mymodule/model"
	"mymodule/utils"
	"os"
)

// Update 用 newData（前端 form 对应的 SelectList）构造一个新的 Inbound，
// 并替换掉配置里 port==oldPort 的那一条。注意：前端的 Reality/TLS/XTLS 开关
// 最终都会反映到 newData.Security 字段里，后端只需读取这个 Security 即可。
func Update(oldPort int, newData model.SelectList) (bool, string) {
	// 1. 读+解密全量配置
	enc, err := os.ReadFile(utils.AES_File)
	if err != nil {
		return false, fmt.Sprintf("读取配置文件失败: %v", err)
	}
	decBytes, err := utils.Decrypt(string(enc))
	if err != nil {
		return false, fmt.Sprintf("解密失败: %v", err)
	}

	// 2. Unmarshal 到 model.Config
	var cfg model.Config
	if err := json.Unmarshal(decBytes, &cfg); err != nil {
		return false, fmt.Sprintf("解析 JSON 失败: %v", err)
	}

	// 3. 构造新的 Inbound
	in := model.Inbound{
		Remark:   newData.Remark,
		Listen:   "",
		Port:     newData.Port,
		Protocol: newData.Protocol,
	}

	// 3.1 填充 Clients
	in.Settings.Clients = []model.Client{
		{
			AlterID:  newData.AlterId,
			ID:       newData.UUID,
			Security: newData.Security,
		},
	}

	// 3.2 初始化 StreamSettings
	in.StreamSettings = &model.StreamSettings{
		Network:  newData.Transport,
		Security: newData.Security,
		TLSSettings: &model.TLSSettings{
			Certificates: []model.Certificate{},
		},
		WSSettings: &model.WSSettings{
			Headers: make(map[string]string),
		},
	}

	// TLS/XTLS 配置
	if newData.Security == "tls" || newData.Security == "xtls" {
		in.StreamSettings.TLSSettings.Certificates = []model.Certificate{
			{
				CertificateFile: newData.CertPath,
				KeyFile:         newData.KeyPath,
			},
		}
	}

	// WebSocket 设置
	if newData.Transport == "ws" {
		in.StreamSettings.WSSettings.Path = newData.WSPath
		in.StreamSettings.WSSettings.Headers["Host"] = newData.SNI
	}

	// 4. 替换旧 inbound
	replaced := false
	for i, inv := range cfg.Inbounds {
		if inv.Port == oldPort {
			cfg.Inbounds[i] = in
			replaced = true
			break
		}
	}
	if !replaced {
		return false, fmt.Sprintf("未找到端口 %d 对应的 inbound", oldPort)
	}

	// 5. 序列化 → 加密 → 写入
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, fmt.Sprintf("JSON 序列化失败: %v", err)
	}
	cipher, err := utils.Encrypt(out)
	if err != nil {
		return false, fmt.Sprintf("加密失败: %v", err)
	}
	if err := os.WriteFile(utils.AES_File, []byte(cipher), 0o600); err != nil {
		return false, fmt.Sprintf("写入配置文件失败: %v", err)
	}

	return true, ""
}
