package helper

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"tugas1-go/pertemuan-10-api-mongodb/app/model"
)

var portfolioValidator = newPortfolioValidator()

func newPortfolioValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return v
}

// DecodePortfolioRequest accepts only the declared JSON fields and one JSON value.
func DecodePortfolioRequest(body []byte) (model.PortfolioRequest, error) {
	var request model.PortfolioRequest
	if err := decodeJSON(body, &request); err != nil {
		return model.PortfolioRequest{}, err
	}
	return request, nil
}

// DecodeProjectRequest accepts a typed project body without database operators.
func DecodeProjectRequest(body []byte) (model.ProjectRequest, error) {
	var request model.ProjectRequest
	if err := decodeJSON(body, &request); err != nil {
		return model.ProjectRequest{}, err
	}
	return request, nil
}

func decodeJSON(body []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			return io.ErrUnexpectedEOF
		}
		return err
	}
	return nil
}

// ValidatePortfolioRequest checks the complete replacement payload, including its arrays.
func ValidatePortfolioRequest(request model.PortfolioRequest) map[string]string {
	fields := make(map[string]string)
	if strings.TrimSpace(request.Headline) == "" {
		fields["headline"] = "wajib diisi"
	}
	if strings.TrimSpace(request.Bio) == "" {
		fields["bio"] = "wajib diisi"
	}
	if request.Projects == nil {
		fields["projects"] = "wajib berupa array"
	}
	if request.Certificates == nil {
		fields["certificates"] = "wajib berupa array"
	}
	for _, project := range request.Projects {
		if strings.TrimSpace(project.Title) == "" {
			fields["projects"] = "setiap project wajib memiliki title"
		}
		for _, technology := range project.Technologies {
			if strings.TrimSpace(technology) == "" {
				fields["projects"] = "technology tidak boleh kosong"
			}
		}
	}
	for _, certificate := range request.Certificates {
		if strings.TrimSpace(certificate.Name) == "" || strings.TrimSpace(certificate.Issuer) == "" {
			fields["certificates"] = "nama dan penerbit sertifikat wajib diisi"
		}
	}
	if err := portfolioValidator.Struct(request); err != nil {
		for _, validationErr := range err.(validator.ValidationErrors) {
			field := validationErr.Field()
			if _, exists := fields[field]; !exists {
				fields[field] = "format tidak valid"
			}
		}
	}
	return fields
}

// ValidateProjectRequest checks a single project before it reaches the repository.
func ValidateProjectRequest(request model.ProjectRequest) map[string]string {
	fields := make(map[string]string)
	if strings.TrimSpace(request.Title) == "" {
		fields["title"] = "wajib diisi"
	}
	for _, technology := range request.Technologies {
		if strings.TrimSpace(technology) == "" {
			fields["technologies"] = "technology tidak boleh kosong"
		}
	}
	if err := portfolioValidator.Struct(request); err != nil {
		for _, validationErr := range err.(validator.ValidationErrors) {
			fields[validationErr.Field()] = "format tidak valid"
		}
	}
	return fields
}
