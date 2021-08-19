package routes

import (
	"notepad/controllers"
	"notepad/middleware"

	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2/middleware/session"
)

func Setup(app *fiber.App) {
	v1 := app.Group("v1")

	// Get local
	v1.Get("secret/userinfo", middleware.SetUserInfo)
	v1.Get("secret/logout", middleware.Logout)
	// Get public
	v1.Get("public/contents-all:srot?", controllers.GetContentsAllPublic)
	v1.Get("public/tags-all", controllers.GetTagsPublic)
	v1.Get("public/:slug", controllers.GetContentsPublic)
	v1.Get("public/related/:slug", controllers.GetRelatedPublic)
	// Get local use middleware
	auth := v1.Use(middleware.IsAuthenticate)
	auth.Get("projects-list", controllers.GetProjects)
	auth.Get("pages/:project/", controllers.GetContentsAllLocal)
	auth.Get("pages/:project/:slug", controllers.GetContentsLocal)
	auth.Get("related/:project/:slug", controllers.GetRelatedLocal)
	auth.Get("tags/:project", controllers.GetTagsLocal)

	// get images
	v1.Get("images/:project/all", controllers.GetImages)

	// Post
	// auth.Post("get-content", controllers.GetContent)
	// auth.Post("get-contents-all", controllers.GetContentsAll)
	v1.Post("create-table", controllers.CreateTable)
	v1.Post("post-content", controllers.PostContent)
	v1.Post("get-related-only", controllers.GetRelatedOnly)
	v1.Post("delete-content", controllers.DeleteContent)

	auth.Post("images/upload", controllers.UploadImage)
	auth.Delete("images/:project/:filename", controllers.DeleteImage)
	auth.Delete("projects/:project", controllers.DeleteProject)
	// app.Get("/", func(c *fiber.Ctx) error {
	// 	return c.SendString("Hello Fiber!")
	// })
}
