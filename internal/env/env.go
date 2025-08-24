package env

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

// Init读取.env文件
func Init(envFiles ...string) error {
	if len(envFiles) == 0 {
		envFiles = []string{".env"}
	}
	return godotenv.Load(envFiles...)
}

func GetString(key string, fallback string) string {
	// 找环境变量里面的Key
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}
