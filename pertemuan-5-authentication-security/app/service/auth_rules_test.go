package service

import (
	"testing"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

func TestValidateRegisterValid(t *testing.T) {
	if errs := ValidateRegister(model.RegisterRequest{Username: " Falah_01 ", Email: "falah@example.com", Password: "Strong123"}); len(errs) != 0 {
		t.Fatalf("expected valid registration, got %v", errs)
	}
}

func TestValidateRegisterRejectsUsername(t *testing.T) {
	cases := []struct{ name, username, field string }{{"blank", " ", "username"}, {"short", "ab", "username"}, {"characters", "bad-name", "username"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := ValidateRegister(model.RegisterRequest{Username: tc.username, Email: "a@b.com", Password: "Strong123"})[tc.field]; !ok {
				t.Fatalf("expected %s error", tc.field)
			}
		})
	}
}

func TestValidateRegisterRejectsEmailAndPasswords(t *testing.T) {
	cases := []struct{ name, email, password string }{{"email", "not-email", "Strong123"}, {"short", "a@b.com", "Abc123"}, {"no digit", "a@b.com", "Password"}, {"no letter", "a@b.com", "12345678"}, {"weak", "a@b.com", "password1"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateRegister(model.RegisterRequest{Username: "user", Email: tc.email, Password: tc.password})
			if len(errs) == 0 {
				t.Fatal("expected validation errors")
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	if errs := ValidateLogin(model.LoginRequest{Username: "user", Password: "old"}); len(errs) != 0 {
		t.Fatalf("expected valid login, got %v", errs)
	}
	if _, ok := ValidateLogin(model.LoginRequest{Password: "old"})["username"]; !ok {
		t.Fatal("expected username error")
	}
	if _, ok := ValidateLogin(model.LoginRequest{Username: "user"})["password"]; !ok {
		t.Fatal("expected password error")
	}
}
