package route

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"tugas1-go/uts-pbl-siakad-mini/app/handler"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
)

type routeAuthStore struct{ user model.User }

func (s routeAuthStore) FindUserByEmail(context.Context, string) (model.User, error) {
	return s.user, nil
}
func (s routeAuthStore) FindActiveAccount(_ context.Context, id int64) (model.CurrentUser, error) {
	return model.CurrentUser{ID: id, Email: s.user.Email, Role: s.user.Role}, nil
}

func TestRegisterOnlyAuthEndpoints(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password-123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	store := routeAuthStore{user: model.User{ID: 1, Email: "admin@example.com", Password: string(hash), Role: "admin"}}
	authService := service.NewAuthService(store, "route-test-secret-long-enough", time.Hour)
	authHandler := handler.NewAuthHandler(authService, store)
	app := fiber.New()
	Register(app, authHandler, []byte("route-test-secret-long-enough"), store)
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/me"} {
		request := httptest.NewRequest("OPTIONS", path, nil)
		response, err := app.Test(request, -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode == 404 {
			t.Fatalf("route %s is missing", path)
		}
	}
	request := httptest.NewRequest("GET", "/api/v1/students", nil)
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatalf("unexpected non-auth endpoint status: %d", response.StatusCode)
	}
}
