package helper

import (
	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
)

const LocalsAuthUser = "auth_user"

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
