package routes

import (
	"notepad/controllers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("api")

	// Post
	api.Get("show", controllers.Show)
	api.Post("post", controllers.Post)
	api.Post("post-content", controllers.PostContent)
	api.Post("get-content", controllers.GetContent)
	api.Post("get-contents-all", controllers.GetContentsAll)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Fiber!")
	})
}
