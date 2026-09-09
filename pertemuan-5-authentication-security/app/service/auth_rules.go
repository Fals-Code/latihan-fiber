package service

import (
	"net/mail"
	"strings"
	"unicode"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

var weakPasswords = map[string]bool{
	"password1":   true,
	"12345678":    true,
	"qwerty123":   true,
	"admin123":    true,
	"password123": true,
}

func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" {
		errs["username"] = "wajib diisi"
	} else if len([]rune(req.Username)) < 3 {
		errs["username"] = "minimal 3 karakter"
	} else if !validUsername(req.Username) {
		errs["username"] = "hanya boleh berisi huruf, angka, titik, dan underscore"
	}
	if req.Email == "" {
		errs["email"] = "wajib diisi"
	} else if _, err := mail.ParseAddress(req.Email); err != nil || !strings.Contains(req.Email, "@") {
		errs["email"] = "format email tidak valid"
	}
	if len([]rune(req.Password)) < 8 {
		errs["password"] = "minimal 8 karakter"
	} else if weakPasswords[strings.ToLower(req.Password)] {
		errs["password"] = "password terlalu lemah"
	} else if !containsLetter(req.Password) {
		errs["password"] = "harus mengandung huruf"
	} else if !containsDigit(req.Password) {
		errs["password"] = "harus mengandung angka"
	}
	return errs
}

func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.Username) == "" {
		errs["username"] = "wajib diisi"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	}
	return errs
}

func validUsername(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func containsLetter(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func containsDigit(value string) bool {
	for _, r := range value {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
