// selectservice/select.go
package selectservice

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"mymodule/model"
	"mymodule/utils"
)

var (
	once      sync.Once
	cacheList []model.SelectList
	cacheErr  error
)

// ResetCache 清空并重新加载缓存
func ResetCache() {
	// 重置 once，使下一次调用 GetByPortWithCache 时重新读取配置文件
	once = sync.Once{}
	cacheList = nil
	cacheErr = nil
}

// loadCache 只执行一次，填充 cacheList
func loadCache() {
	cacheList, cacheErr = Select()
}

// GetByPortWithCache 缓存并返回特定端口的 SelectList
func GetByPortWithCache(port int) (model.SelectList, error) {
	once.Do(loadCache)
	if cacheErr != nil {
		return model.SelectList{}, cacheErr
	}
	for _, it := range cacheList {
		if it.Port == port {
			return it, nil
		}
	}
	return model.SelectList{}, fmt.Errorf("未找到端口 %d 的配置", port)
}

// Select 读取、解密配置，返回前端需要的 SelectList 列表
func Select() ([]model.SelectList, error) {
	// 1. 读取文件
	enc, err := os.ReadFile(utils.AES_File)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 2. 解密
	dec, err := utils.Decrypt(string(enc))
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	// 3. 打印解密后的 JSON，而不是 Go 结构体
	log.Println("解密后的完整配置 JSON:")
	log.Println(string(dec))

	// 4. 反序列化到 Config
	var cfg model.Config
	if err := json.Unmarshal(dec, &cfg); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	// 5. 遍历 Inbounds，构造 SelectList
	var list []model.SelectList
	for _, ib := range cfg.Inbounds {
		// 跳过无效条目
		if ib.Port <= 0 || ib.StreamSettings == nil || len(ib.Settings.Clients) == 0 {
			continue
		}
		cli := ib.Settings.Clients[0]
		ss := ib.StreamSettings

		item := model.SelectList{
			Remark:     ib.Remark,
			Port:       ib.Port,
			Protocol:   ib.Protocol,
			UUID:       cli.ID,
			AlterId:    cli.AlterID,
			CipherVM:   cli.Security,
			Transport:  ss.Network,
			Security:   ss.Security,
			TLSVM:      ss.Security == "tls",
			Sniffing:   false,
			Fallback:   "",
			WSPath:     "",
			SNI:        "",
			CertPath:   "",
			KeyPath:    "",
			MinVersion: "",
			MaxVersion: "",
			Cipher:     "",
			Domain:     "",
			Alpn:       "",
			Renew:      0,
			Reality:    false,
			Show:       false,
			Xver:       0,
			ServerAddr: "",
			PublicKey:  "",
			PrivateKey: "",
			ShortIds:   "",
		}

		// Fallback
		if len(ib.Fallbacks) > 0 {
			item.Fallback = ib.Fallbacks[0].Path
		}

		// Sniffing
		if ib.Sniffing != nil && ib.Sniffing.Enabled {
			item.Sniffing = true
			item.SniffingOpts = ib.Sniffing.DestOverride
		}

		// TLS/XTLS 证书 & 参数
		if ss.TLSSettings != nil {
			if len(ss.TLSSettings.Certificates) > 0 {
				item.CertPath = ss.TLSSettings.Certificates[0].CertificateFile
				item.KeyPath = ss.TLSSettings.Certificates[0].KeyFile
			}
			item.MinVersion = ss.TLSSettings.MinVersion
			item.MaxVersion = ss.TLSSettings.MaxVersion
			item.Cipher = ss.TLSSettings.Cipher
			item.Domain = ss.TLSSettings.ServerName
			item.Alpn = strings.Join(ss.TLSSettings.Alpn, ",")
			item.Renew = ss.TLSSettings.Renew
		}

		// WebSocket
		if ss.WSSettings != nil {
			item.WSPath = ss.WSSettings.Path
			if host, ok := ss.WSSettings.Headers["Host"]; ok {
				item.SNI = host
			}
		}

		// Reality
		if ss.Security == "reality" && ss.RealitySettings != nil {
			rs := ss.RealitySettings
			item.Reality = true
			item.Show = rs.Show
			item.Xver = rs.Xver
			item.ServerAddr = rs.ServerName
			item.PublicKey = rs.PublicKey
			item.PrivateKey = rs.PrivateKey
			item.ShortIds = rs.ShortID
			item.MinVersion = rs.MinVersion
			item.MaxVersion = rs.MaxVersion
			item.Alpn = strings.Join(rs.Alpn, ",")
		}

		list = append(list, item)
	}

	return list, nil
}
