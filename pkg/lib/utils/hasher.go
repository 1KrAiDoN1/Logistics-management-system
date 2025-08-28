package utils

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func HashPassword(Password string) (string, error) {
	hash := sha1.New()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	secretString := os.Getenv("SECRET_HASH") // получаем значение из файла конфигурации
	_, err1 := hash.Write([]byte(Password))
	if err1 != nil {
		log.Println("Ошибка при шифровании пароля", err)
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum([]byte(secretString))), nil

}
