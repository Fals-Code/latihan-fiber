package helper

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var requestValidator = newValidator()

var weakPasswords = map[string]bool{
	"password1":   true,
	"12345678":    true,
	"qwerty123":   true,
	"admin123":    true,
	"password123": true,
}

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		return strings.Split(field.Tag.Get("json"), ",")[0]
	})
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}
		if weakPasswords[strings.ToLower(value)] {
			return false
		}
		var letter, digit bool
		for _, r := range value {
			letter = letter || unicode.IsLetter(r)
			digit = digit || unicode.IsDigit(r)
		}
		return letter && digit
	})
	_ = v.RegisterValidation("studentnim", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(int)
		return ok && value > 0
	})
	return v
}

func ValidateRequest(request any) map[string]string {
	errs := requestValidator.Struct(request)
	if errs == nil {
		return nil
	}

	result := make(map[string]string)
	for _, err := range errs.(validator.ValidationErrors) {
		field := err.Field()
		if field == "" {
			field = jsonFieldName(request, err.StructField())
		}
		result[field] = validationMessage(err)
	}
	return result
}

func jsonFieldName(request any, field string) string {
	typ := reflect.TypeOf(request)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if structField, ok := typ.FieldByName(field); ok {
		name := strings.Split(structField.Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return field
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "wajib diisi"
	case "gt":
		return "harus lebih besar dari 0"
	case "gte":
		return "harus lebih besar atau sama dengan 0"
	case "lte":
		return "harus kurang dari atau sama dengan 100"
	case "min":
		return "minimal " + err.Param() + " karakter"
	case "max":
		return "maksimal " + err.Param() + " karakter"
	case "strongpassword":
		if value, ok := err.Value().(string); ok && weakPasswords[strings.ToLower(value)] {
			return "password terlalu umum"
		}
		return "harus mengandung huruf dan angka"
	case "studentnim":
		return "NIM harus berupa angka positif"
	default:
		return "format tidak valid"
	}
}
