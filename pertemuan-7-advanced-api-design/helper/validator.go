package helper

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var requestValidator = validator.New()

func ValidateRequest(request any) map[string]string {
	errs := requestValidator.Struct(request)
	if errs == nil {
		return nil
	}

	result := make(map[string]string)
	for _, err := range errs.(validator.ValidationErrors) {
		field := jsonFieldName(request, err.StructField())
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
	default:
		return "format tidak valid"
	}
}
