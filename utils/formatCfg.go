// utils/formatter.go
package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"mymodule/model"
)

// FormatCfg 接收前端 raw JSON，将其合并到完整 Config 中，
// 清理所有脏数据，然后读写磁盘上的加密文件，最后返回明文 JSON。
func FormatCfg(raw []byte) ([]byte, error) {
	// 1. 解析前端表单
	var in struct {
		Remark       string   `json:"Remark"`
		Port         int      `json:"Port"`
		Protocol     string   `json:"Protocol"`
		UUID         string   `json:"UUID"`
		AlterId      int      `json:"AlterId"`
		CipherVM     string   `json:"CipherVM"`
		Transport    string   `json:"Transport"`
		WSPath       string   `json:"WSPath"`
		Fallback     string   `json:"Fallback"`
		Reality      bool     `json:"Reality"`
		TLS          bool     `json:"TLS"`
		XTLS         bool     `json:"XTLS"`
		TLSVM        bool     `json:"TLSVM"`
		CertPath     string   `json:"CertPath"`
		KeyPath      string   `json:"KeyPath"`
		MinVersion   string   `json:"MinVersion"`
		MaxVersion   string   `json:"MaxVersion"`
		Cipher       string   `json:"Cipher"`
		Domain       string   `json:"Domain"`
		Alpn         string   `json:"Alpn"`
		Renew        int      `json:"Renew"`
		Sniffing     bool     `json:"Sniffing"`
		SniffingOpts []string `json:"SniffingOpts"`
		Show         bool     `json:"Show"`
		Xver         int      `json:"Xver"`
		ServerAddr   string   `json:"ServerAddr"`
		PrivateKey   string   `json:"PrivateKey"`
		PublicKey    string   `json:"PublicKey"`
		ShortIds     string   `json:"ShortIds"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("解析前端 JSON 失败: %w", err)
	}

	// 2. 构造新的 Inbound
	ib := model.Inbound{
		Remark:   in.Remark,
		Listen:   "0.0.0.0",
		Port:     in.Port,
		Protocol: in.Protocol,
		Settings: model.InboundSettings{
			Clients: []model.Client{{
				AlterID: in.AlterId,
				ID:      in.UUID,
				Security: func() string {
					if in.Protocol == "vmess" {
						return in.CipherVM
					}
					return in.Cipher
				}(),
			}},
		},
		StreamSettings: &model.StreamSettings{
			Network: in.Transport,
			TLSSettings: &model.TLSSettings{
				Certificates: []model.Certificate{},
			},
			WSSettings: &model.WSSettings{
				Headers: make(map[string]string),
			},
		},
	}

	// 3. VMess 回落（Fallback）
	if in.Protocol == "vmess" && in.Fallback != "" {
		ib.Fallbacks = []model.Fallback{{Path: in.Fallback}}
	}

	// 4. 嗅探配置
	if in.Sniffing {
		ib.Sniffing = &model.Sniffing{
			Enabled:      true,
			DestOverride: in.SniffingOpts,
		}
	}

	// 5. 安全层：Reality / XTLS / TLS
	switch {
	case in.Reality:
		ib.StreamSettings.Security = "reality"
		ib.StreamSettings.RealitySettings = &model.RealitySettings{
			Show:        in.Show,
			Fingerprint: in.Cipher,
			ServerName:  in.ServerAddr,
			PublicKey:   in.PublicKey,
			PrivateKey:  in.PrivateKey,
			ShortID:     in.ShortIds,
			Xver:        in.Xver,
			MinVersion:  in.MinVersion,
			MaxVersion:  in.MaxVersion,
			Alpn:        strings.Split(in.Alpn, ","),
		}
	case in.XTLS:
		ib.StreamSettings.Security = "xtls"
		ib.StreamSettings.TLSSettings.Certificates = []model.Certificate{{
			CertificateFile: in.CertPath,
			KeyFile:         in.KeyPath,
		}}
		ib.StreamSettings.TLSSettings.MinVersion = in.MinVersion
		ib.StreamSettings.TLSSettings.MaxVersion = in.MaxVersion
		ib.StreamSettings.TLSSettings.Cipher = in.Cipher
		ib.StreamSettings.TLSSettings.ServerName = in.Domain
		ib.StreamSettings.TLSSettings.Alpn = strings.Split(in.Alpn, ",")
		ib.StreamSettings.TLSSettings.Renew = in.Renew
	case in.TLS || in.TLSVM:
		ib.StreamSettings.Security = "tls"
		ib.StreamSettings.TLSSettings.Certificates = []model.Certificate{{
			CertificateFile: in.CertPath,
			KeyFile:         in.KeyPath,
		}}
		ib.StreamSettings.TLSSettings.MinVersion = in.MinVersion
		ib.StreamSettings.TLSSettings.MaxVersion = in.MaxVersion
		ib.StreamSettings.TLSSettings.Cipher = in.Cipher
		ib.StreamSettings.TLSSettings.ServerName = in.Domain
		ib.StreamSettings.TLSSettings.Alpn = strings.Split(in.Alpn, ",")
		ib.StreamSettings.TLSSettings.Renew = in.Renew
	default:
		ib.StreamSettings.Security = ""
	}

	// 6. WebSocket
	if in.Transport == "ws" {
		ib.StreamSettings.WSSettings.Path = in.WSPath
		ib.StreamSettings.WSSettings.Headers["Host"] = in.Domain
	}

	// 7. 合并到完整 Config
	exist, err := FileExists(AES_File)
	if err != nil {
		return nil, fmt.Errorf("检查配置文件失败: %w", err)
	}

	var cfg model.Config
	if !exist {
		// 文件不存在：新建
		cfg = model.Config{
			Log:       model.Log{LogLevel: "none"},
			DNS:       model.DNS{Servers: []string{"1.1.1.1", "8.8.8.8", "1.0.0.1", "8.8.4.4"}},
			Outbounds: []model.Outbound{{Protocol: "freedom", Settings: map[string]interface{}{}}},
			Inbounds:  []model.Inbound{ib},
		}
	} else {
		// 文件存在：读取、解密、反序列化
		data, err := os.ReadFile(AES_File)
		if err != nil {
			return nil, fmt.Errorf("读取旧配置失败: %w", err)
		}
		dec, err := Decrypt(string(data))
		if err != nil {
			return nil, fmt.Errorf("解密旧配置失败: %w", err)
		}
		if err := json.Unmarshal(dec, &cfg); err != nil {
			return nil, fmt.Errorf("解析旧配置失败: %w", err)
		}

		// 清理脏数据：剔除 Port<=0、无 Clients、无 StreamSettings
		clean := cfg.Inbounds[:0]
		for _, old := range cfg.Inbounds {
			if old.Port > 0 && old.StreamSettings != nil && len(old.Settings.Clients) > 0 {
				clean = append(clean, old)
			}
		}
		cfg.Inbounds = clean

		// 替换或追加
		replaced := false
		for i := range cfg.Inbounds {
			if cfg.Inbounds[i].Port == ib.Port {
				cfg.Inbounds[i] = ib
				replaced = true
				break
			}
		}
		if !replaced {
			cfg.Inbounds = append(cfg.Inbounds, ib)
		}
	}

	// 8. 序列化、加密、写回
	plain, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化 Config 失败: %w", err)
	}
	enc, err := Encrypt(plain)
	if err != nil {
		return nil, fmt.Errorf("加密 Config 失败: %w", err)
	}
	if err := ioutil.WriteFile(AES_File, []byte(enc), 0644); err != nil {
		return nil, fmt.Errorf("写入配置文件失败: %w", err)
	}

	return plain, nil
}
