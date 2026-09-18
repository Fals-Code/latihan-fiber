package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-6-authorization-rbac/app/model"
	"tugas1-go/pertemuan-6-authorization-rbac/helper"
)

func TestRequirePermission(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"user:list", "user:update:any"},
		"staff": {"user:list"},
		"user":  {},
	})
	for _, tc := range []struct {
		name       string
		identity   *model.AuthUser
		permission string
		status     int
	}{
		{"missing identity", nil, "user:list", fiber.StatusUnauthorized},
		{"allowed", &model.AuthUser{UserID: 1, Role: "admin"}, "user:list", fiber.StatusNoContent},
		{"missing permission", &model.AuthUser{UserID: 1, Role: "user"}, "user:list", fiber.StatusForbidden},
		{"unknown role", &model.AuthUser{UserID: 1, Role: "unknown"}, "user:list", fiber.StatusForbidden},
		{"unknown permission", &model.AuthUser{UserID: 1, Role: "admin"}, "user:unknown", fiber.StatusForbidden},
		{"admin can create users", &model.AuthUser{UserID: 1, Role: "admin"}, "user:update:any", fiber.StatusNoContent},
		{"staff cannot create users", &model.AuthUser{UserID: 1, Role: "staff"}, "user:update:any", fiber.StatusForbidden},
		{"user cannot create users", &model.AuthUser{UserID: 1, Role: "user"}, "user:update:any", fiber.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				if tc.identity != nil {
					c.Locals(helper.LocalsAuthUser, *tc.identity)
				}
				return c.Next()
			}, RequirePermission(perms, tc.permission), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusNoContent)
			})
			resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status {
				t.Fatalf("got %d, want %d", resp.StatusCode, tc.status)
			}
		})
	}
}

func TestRequirePermissionNilSetDenies(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals(helper.LocalsAuthUser, model.AuthUser{UserID: 1, Role: "admin"})
		return c.Next()
	}, RequirePermission(nil, "user:list"))
	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("got %d", resp.StatusCode)
	}
}
