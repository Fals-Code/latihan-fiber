package helper

import (
	"encoding/csv"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

func TestNegotiateRejectsUnsupportedAccept(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		appErr, ok := err.(*AppError)
		if !ok {
			return err
		}
		return c.SendStatus(appErr.Status)
	}})
	app.Get("/", func(c *fiber.Ctx) error { return Negotiate(c, 200, "ok", nil, nil) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/plain")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != fiber.StatusNotAcceptable {
		t.Fatalf("expected 406, got status=%d err=%v", resp.StatusCode, err)
	}
}

func TestToCSVSerializesOwnerPointers(t *testing.T) {
	ownerID := 42
	data := []model.Student{{ID: 1, NIM: 900001, Name: "Owned", OwnerID: &ownerID}, {ID: 2, NIM: 900002, Name: "Unowned"}}
	output := toCSV(data)
	if strings.Contains(output, "0x") {
		t.Fatalf("csv contains pointer address: %q", output)
	}
	records, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatalf("csv parse failed: %v", err)
	}
	if len(records) != 3 || len(records[0]) != 7 {
		t.Fatalf("unexpected csv shape: %#v", records)
	}
	if records[1][5] != "42" || records[2][5] != "" {
		t.Fatalf("unexpected owner values: %#v", records)
	}
	for _, record := range records {
		if len(record) != 7 {
			t.Fatalf("inconsistent column count: %#v", records)
		}
	}
}

func TestStudentJSONPreservesOwnerID(t *testing.T) {
	ownerID := 42
	encoded, err := json.Marshal(model.Student{ID: 1, OwnerID: &ownerID})
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	if !strings.Contains(string(encoded), `"owner_id":42`) {
		t.Fatalf("owner id missing from json: %s", encoded)
	}
}

func TestNegotiateCSV(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return Negotiate(c, 200, "ok", []model.Student{{ID: 1, Name: "A"}}, nil) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("request failed: %v", err)
	}
	if got := resp.Header.Get("Content-Type"); got == "" {
		t.Fatal("missing content type")
	}
}
