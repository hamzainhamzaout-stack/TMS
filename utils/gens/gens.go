package gens

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/gomail.v2"
)

func ToVerboseName(s string) string {
	parts := strings.Split(s, "_")

	for i, part := range parts {
		parts[i] = cases.Title(language.English).String(strings.ToLower(part))
	}

	return strings.Join(parts, " ")
}

func GenerateRandomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result strings.Builder
	charsetLength := big.NewInt(int64(len(charset)))

	for range 16 {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			panic(err)
		}
		result.WriteByte(charset[num.Int64()])
	}
	return result.String()
}

func GenerateRandomPassword() (string, error) {
	minLength := 8
	maxLength := 16
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$&"
	var password strings.Builder
	charsetLength := big.NewInt(int64(len(charset)))

	length, err := rand.Int(rand.Reader, big.NewInt(int64(maxLength-minLength+1)))
	if err != nil {
		return "", err
	}
	passwordLength := int(length.Int64()) + minLength

	for i := 0; i < passwordLength; i++ {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		password.WriteByte(charset[num.Int64()])
	}

	return password.String(), nil
}


func SendEmail(to, subject, body string) (int, error) {
	MAIL_HOST := os.Getenv("MAIL_HOST")
	MAIL_PORT := os.Getenv("MAIL_PORT")
	MAIL_USER := os.Getenv("MAIL_USER")
	MAIL_PASSWORD := os.Getenv("MAIL_PASSWORD")
	FROM_EMAIL := os.Getenv("FROM_EMAIL")
	m := gomail.NewMessage()
	m.SetHeader("From", FROM_EMAIL)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	log.Printf("Mail port type %T", MAIL_PORT)

	port, err := strconv.Atoi(MAIL_PORT)
	if err != nil {
		log.Printf("Invalid MAIL_PORT: %v", err)
		return http.StatusInternalServerError, fmt.Errorf("invalid MAIL PORT: %v", err)
	}

	d := gomail.NewDialer(MAIL_HOST, port, MAIL_USER, MAIL_PASSWORD)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error sending email: %v", err)
		return http.StatusInternalServerError, fmt.Errorf("error sending email: %v", err)
	}

	return http.StatusOK, nil
}

