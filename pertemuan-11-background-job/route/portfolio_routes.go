package route

import (
	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-11-background-job/app/service"
	"tugas1-go/pertemuan-11-background-job/helper"
	middleware "tugas1-go/pertemuan-11-background-job/middleware"
)

// RegisterPortfolioRoutes attaches the portfolio API beneath the existing /api/v1 group.
func RegisterPortfolioRoutes(api fiber.Router, portfolios *service.PortfolioService, jwt *helper.JWTManager) {
	requireAuth := middleware.RequireAuth(jwt)
	students := api.Group("/students", middleware.RequireJSON, requireAuth)
	students.Get("/:id/portfolio", portfolios.Get)
	students.Put("/:id/portfolio", portfolios.Replace)
	students.Post("/:id/portfolio/projects", portfolios.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", portfolios.DeleteProject)

	api.Get("/portfolios/stats/technologies", requireAuth, portfolios.TechnologyStats)
}
