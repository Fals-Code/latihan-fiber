package route

import (
	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-10-api-mongodb/app/service"
	p7helper "tugas1-go/pertemuan-7-advanced-api-design/helper"
	p7middleware "tugas1-go/pertemuan-7-advanced-api-design/middleware"
)

// RegisterPortfolioRoutes attaches the portfolio API beneath the existing /api/v1 group.
func RegisterPortfolioRoutes(api fiber.Router, portfolios *service.PortfolioService, jwt *p7helper.JWTManager) {
	requireAuth := p7middleware.RequireAuth(jwt)
	students := api.Group("/students", p7middleware.RequireJSON, requireAuth)
	students.Get("/:id/portfolio", portfolios.Get)
	students.Put("/:id/portfolio", portfolios.Replace)
	students.Post("/:id/portfolio/projects", portfolios.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", portfolios.DeleteProject)

	api.Get("/portfolios/stats/technologies", requireAuth, portfolios.TechnologyStats)
}
