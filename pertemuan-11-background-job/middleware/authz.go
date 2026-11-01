package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-11-background-job/helper"
)

func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if !perms.Can(user.Role, permission) {
			return helper.NewAppError(fiber.StatusForbidden, fmt.Sprintf("role %s tidak memiliki hak %s", user.Role, permission))
		}
		return c.Next()
	}
}
