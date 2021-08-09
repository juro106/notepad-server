package routes

import (
	"notepad/controllers"
	"notepad/middleware"

	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2/middleware/session"
)

func Setup(app *fiber.App) {
	v1 := app.Group("v1")
	// Get public
	v1.Get("public/contents-all", controllers.GetPublicContentsAll)
	v1.Get("public/:slug", controllers.GetPublicContents)
	v1.Get("public/related/:slug", controllers.GetPublicRelated)

	// Get local
	v1.Get("secret/userinfo", middleware.SetUserInfo)
	v1.Get("secret/logout", middleware.Logout)
	// Get local use middleware
	auth := v1.Use(middleware.IsAuthenticate)
	auth.Get("projects-list", controllers.GetProjects)
	auth.Get("pages/:projects/", controllers.GetContentsAll)
	auth.Get("pages/:projects/:slug", controllers.GetContents)
	auth.Get("related/:projects/:slug", controllers.GetRelated)

	v1.Get("show", controllers.Show)
	// get images
	v1.Get("images/:project/all", controllers.GetImages)

	// Post
	// auth.Post("get-content", controllers.GetContent)
	// auth.Post("get-contents-all", controllers.GetContentsAll)
	v1.Post("post", controllers.Post)
	v1.Post("create-table", controllers.CreateTable)
	v1.Post("post-content", controllers.PostContent)
	v1.Post("get-related-only", controllers.GetRelatedOnly)
	v1.Post("delete-content", controllers.DeleteContent)

	auth.Post("images/upload", controllers.UploadImage)
	auth.Delete("images/:projects/:filename", controllers.DeleteImage)
	// app.Get("/", func(c *fiber.Ctx) error {
	// 	return c.SendString("Hello Fiber!")
	// })
}
