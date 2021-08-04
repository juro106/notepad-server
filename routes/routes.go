package routes

import (
	"notepad/controllers"
	"notepad/middleware"

	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2/middleware/session"
)

func Setup(app *fiber.App) {
	// api := app.Group("api")

	app.Get("secret/userinfo", middleware.SetUserInfo)
	auth := app.Use(middleware.IsAuthenticate)
	auth.Get("projects-list", controllers.GetProjects)
	auth.Get(":projects/", controllers.GetContentsAll)
	auth.Get(":projects/:slug", controllers.GetContents)

	auth.Post("get-content", controllers.GetContent)
	auth.Post("get-contents-all", controllers.GetContentsAll)
	// Get
	app.Get("show", controllers.Show)
	// app.Get(":projects/:slug", controller.GetContentN)

	// middleware
	// app.Get("secret/userinfo", middleware.SetUserInfo)
	// auth.Get("secret/uinfo", controllers.SecretUserInfo)

	// Post
	app.Post("post", controllers.Post)
	app.Post("create-table", controllers.CreateTable)
	app.Post("post-content", controllers.PostContent)
	// app.Post("get-content", controllers.GetContent)
	// app.Post("get-contents-all", controllers.GetContentsAll)
	app.Post("get-related", controllers.GetRelated)
	app.Post("get-related-only", controllers.GetRelatedOnly)
	app.Post("delete-content", controllers.DeleteContent)

	// app.Get("/", func(c *fiber.Ctx) error {
	// 	return c.SendString("Hello Fiber!")
	// })
}
