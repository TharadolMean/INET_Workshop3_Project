package routes

import (
	"employee-profile-api/handlers"
	"employee-profile-api/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func NewApp(db *gorm.DB, username, password string) *fiber.App {
	app := fiber.New(fiber.Config{UnescapePath: true, ErrorHandler: func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		return c.Status(code).JSON(fiber.Map{"error": "request failed"})
	}})
	api := app.Group("/api/v1")
	users := api.Group("/users")
	h := handlers.EmployeeHandler{DB: db}
	auth := middleware.BasicAuth(username, password)
	users.Get("/generations", h.Generations)
	users.Get("/search/:keyword", h.Search)
	users.Get("/", h.List)
	users.Get("/:id", h.Get)
	users.Post("/", auth, h.Create)
	users.Put("/:id", auth, h.Update)
	users.Delete("/:id", auth, h.Delete)
	return app
}
