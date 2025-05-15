package usersettingservice

import (
	"encoding/json"
	"errors"
	"log"
	"mymodule/utils"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type config struct {
	User string `json:"user"`
	Pwd  string `json:"pwd"`
}

var cfg config

func init() {
	cfg = loadConfig()
}

func loadConfig() config {
	data, err := os.ReadFile("./config/user.json")
	if err != nil {
		log.Fatal(err)
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg

}

func VerifyOldPwd(pwd string, user string) (bool, string, error) {
	err := bcrypt.CompareHashAndPassword([]byte(cfg.Pwd), []byte(pwd))
	if err != nil {
		return false, "", err
	}
	token, err := utils.GenerateToken(user)
	if err != nil {
		return false, "", err
	}
	return true, token, nil
}

func ReSet(newName string, newPwd string, user string, token string) (bool, error) {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return false, err
	}
	// 再次检查过期时间
	if time.Now().Unix() > claims.ExpiresAt {
		err := errors.New("token 已过期")
		return false, err
	}

	// 校验 Token 中的 Username 与入参 user 是否一致
	if claims.Username != user {
		err := errors.New("无效token！")
		return false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	hashedPwd := string(hash)

	if err := updateConfig(newName, hashedPwd); err != nil {
		return false, err
	}
	return true, nil
}

// updateConfig 读取 path 对应的 JSON 文件，更新 user 和 pwd 字段，写回磁盘
func updateConfig(user, hashedPwd string) error {
	// 1. 读文件
	var path string = "./config/user.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 2. 解析到结构体
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// 3. 更新字段
	cfg.User = user
	cfg.Pwd = hashedPwd

	// 4. Marshal 并写回（带缩进更友好）
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// 先写到临时文件，再原子替换
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
