package tools

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func updateEnvFile(key string, value string) error {

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	envPath := filepath.Join(dir, ".env")

	envMap, err := godotenv.Read(envPath)
	if err != nil {
		return err
	}

	envMap[key] = value

	err = godotenv.Write(envMap, envPath)
	if err != nil {
		return err
	}

	return nil
}
