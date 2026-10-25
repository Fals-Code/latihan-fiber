package helper

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"tugas1-go/pertemuan-10-api-mongodb/app/model"
)

func TestDecodePortfolioRequestRejectsOperatorInStringField(t *testing.T) {
	_, err := DecodePortfolioRequest([]byte(`{"headline":{"$ne":""},"bio":"bio","projects":[],"certificates":[]}`))
	if err == nil {
		t.Fatal("expected an object where a string is required to be rejected")
	}
}

func TestDecodePortfolioRequestRejectsUnknownTopLevelOperator(t *testing.T) {
	_, err := DecodePortfolioRequest([]byte(`{"headline":"Backend","bio":"bio","projects":[],"certificates":[],"$set":{"admin":true}}`))
	if err == nil {
		t.Fatal("expected an unknown MongoDB operator field to be rejected")
	}
}

func TestValidatePortfolioRequestRejectsMoreThanTwentyProjects(t *testing.T) {
	request := validPortfolioRequest()
	for i := 0; i < 21; i++ {
		request.Projects = append(request.Projects, model.ProjectRequest{
			Title:        fmt.Sprintf("Project %d", i),
			Technologies: []string{"Go"},
			Link:         "https://example.com/project",
			Year:         2026,
		})
	}

	if fields := ValidatePortfolioRequest(request); fields["projects"] == "" {
		t.Fatalf("expected projects limit validation, got %v", fields)
	}
}

func TestValidatePortfolioRequestAcceptsTwentyProjects(t *testing.T) {
	request := validPortfolioRequest()
	for i := 0; i < 20; i++ {
		request.Projects = append(request.Projects, model.ProjectRequest{
			Title:        fmt.Sprintf("Project %d", i),
			Technologies: []string{"Go"},
			Link:         "https://example.com/project",
			Year:         2026,
		})
	}

	if fields := ValidatePortfolioRequest(request); len(fields) != 0 {
		t.Fatalf("valid request was rejected: %v", fields)
	}
}

func TestValidatePortfolioRequestRequiresProjectAndCertificateFields(t *testing.T) {
	request := validPortfolioRequest()
	request.Projects = []model.ProjectRequest{{}}
	request.Certificates = []model.CertificateRequest{{}}

	fields := ValidatePortfolioRequest(request)
	for _, field := range []string{"title", "name"} {
		if !hasValidationField(fields, field) {
			t.Errorf("expected validation error for %q, got %v", field, fields)
		}
	}
}

func TestPortfolioValidationRejectsWhitespaceOnlyText(t *testing.T) {
	request := model.PortfolioRequest{
		Headline: "Engineer",
		Bio:      "Bio",
		Projects: []model.ProjectRequest{{
			Title: " ", Technologies: []string{"Go", "  "}, Link: "https://example.test", Year: 2025,
		}},
		Certificates: []model.CertificateRequest{{Name: " ", Issuer: "Issuer", Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}},
	}
	if fields := ValidatePortfolioRequest(request); len(fields) == 0 {
		t.Fatal("expected whitespace-only fields to fail validation")
	}
}

func validPortfolioRequest() model.PortfolioRequest {
	return model.PortfolioRequest{
		Headline:     "Backend Developer",
		Bio:          "Mahasiswa yang mengembangkan API.",
		Projects:     []model.ProjectRequest{},
		Certificates: []model.CertificateRequest{},
	}
}

func hasValidationField(fields map[string]string, want string) bool {
	for field := range fields {
		if field == want || strings.HasPrefix(field, want+".") {
			return true
		}
	}
	return false
}

func TestDecodePortfolioRequestRequiresExactlyOneJSONValue(t *testing.T) {
	body, err := json.Marshal(validPortfolioRequest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodePortfolioRequest(append(body, []byte(` {}`)...)); err == nil {
		t.Fatal("expected trailing JSON value to be rejected")
	}
}
