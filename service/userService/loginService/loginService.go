package loginService

import (
	"encoding/json"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type config struct {
	User string `json:"user"`
	Pwd  string `json:"pwd"`
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
func Login(user string, pwd string) (bool, string) {
	var cfg config = loadConfig()
	if user != cfg.User {
		return false, "用户不存在"
	}

	err := bcrypt.CompareHashAndPassword([]byte(cfg.Pwd), []byte(pwd))
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}
