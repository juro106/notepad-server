package routes

import (
	"notepad/controllers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	// api := app.Group("api")

	// Post
	app.Get("show", controllers.Show)
	app.Post("post", controllers.Post)
	app.Post("post-content", controllers.PostContent)
	app.Post("get-content", controllers.GetContent)
	app.Post("get-contents-all", controllers.GetContentsAll)
	app.Post("get-related", controllers.GetRelated)
	app.Post("get-related-only", controllers.GetRelatedOnly)
	app.Post("delete-content", controllers.DeleteContent)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello Fiber!")
	})
}
