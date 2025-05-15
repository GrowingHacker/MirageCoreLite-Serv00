// model/model.go
package model

// Config 为整份 Xray 配置
type Config struct {
	Log       Log        `json:"log"`
	Outbounds []Outbound `json:"outbounds"`
	DNS       DNS        `json:"dns"`
	Inbounds  []Inbound  `json:"inbounds"`
}

type Log struct {
	LogLevel string `json:"loglevel"`
}

type Outbound struct {
	Protocol string      `json:"protocol"`
	Settings interface{} `json:"settings"`
}

type DNS struct {
	Servers []string `json:"servers"`
}

type Inbound struct {
	Remark         string          `json:"remark"`
	Listen         string          `json:"listen"`
	Port           int             `json:"port"`
	Protocol       string          `json:"protocol"`
	Settings       InboundSettings `json:"settings"`
	StreamSettings *StreamSettings `json:"streamSettings"`
	Sniffing       *Sniffing       `json:"sniffing,omitempty"`
	Fallbacks      []Fallback      `json:"fallbacks,omitempty"`
}

type InboundSettings struct {
	Clients []Client `json:"clients"`
}

type Client struct {
	AlterID  int    `json:"alterId"`
	ID       string `json:"id"`
	Security string `json:"security"`
}

// StreamSettings 传输层设置
type StreamSettings struct {
	Network         string           `json:"network"`
	Security        string           `json:"security"`
	TLSSettings     *TLSSettings     `json:"tlsSettings,omitempty"`
	RealitySettings *RealitySettings `json:"realitySettings,omitempty"`
	WSSettings      *WSSettings      `json:"wsSettings,omitempty"`
	GRPCSettings    *GRPCSettings    `json:"grpcSettings,omitempty"`
}

// TLSSettings TLS/XTLS 证书及参数配置
type TLSSettings struct {
	Certificates []Certificate `json:"certificates"`
	MinVersion   string        `json:"minVersion,omitempty"`
	MaxVersion   string        `json:"maxVersion,omitempty"`
	Cipher       string        `json:"cipher,omitempty"`
	ServerName   string        `json:"serverName,omitempty"`
	Alpn         []string      `json:"alpn,omitempty"`
	Renew        int           `json:"renew,omitempty"`
}

// Certificate 证书文件对
type Certificate struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

// RealitySettings Reality 协议专属配置
type RealitySettings struct {
	Show        bool     `json:"show"`
	Fingerprint string   `json:"fingerprint"`
	ServerName  string   `json:"serverName"`
	PublicKey   string   `json:"publicKey"`
	PrivateKey  string   `json:"privateKey"`
	ShortID     string   `json:"shortId"`
	Xver        int      `json:"xver"`
	MinVersion  string   `json:"minVersion"`
	MaxVersion  string   `json:"maxVersion"`
	Alpn        []string `json:"alpn"`
}

type WSSettings struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

type GRPCSettings struct {
	ServiceName string `json:"serviceName"`
	MultiMode   bool   `json:"multiMode"`
}

type Fallback struct {
	Path     string `json:"path,omitempty"`
	ALPN     string `json:"alpn,omitempty"`
	Dest     string `json:"dest,omitempty"`
	Xver     int    `json:"xver,omitempty"`
	Name     string `json:"name,omitempty"`
	Redirect string `json:"redirect,omitempty"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

// SelectList 用于前端列表展示 & 编辑初始化
type SelectList struct {
	// 基础
	Remark   string
	Port     int
	Protocol string
	UUID     string
	AlterId  int
	CipherVM string

	// 传输
	Transport string
	WSPath    string
	SNI       string
	Fallback  string

	// 安全
	Security string // "", "tls", "xtls", "reality"
	TLSVM    bool
	CertPath string
	KeyPath  string

	// VLESS/TLS 共享
	MinVersion string
	MaxVersion string
	Cipher     string
	Domain     string
	Alpn       string
	Renew      int

	// Reality 专属
	Reality    bool
	Show       bool
	Xver       int
	ServerAddr string
	PublicKey  string
	PrivateKey string
	ShortIds   string

	// Sniffing
	Sniffing     bool
	SniffingOpts []string
}
