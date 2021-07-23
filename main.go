package main

import (
	"log"
	"notepad/database"
	"notepad/routes"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

func main() {
	_, err := database.DbInit()
	if err != nil {
		log.Fatal(err)
	}
	defer database.DbClose()

	app := fiber.New()

	routes.Setup(app)

	app.Listen(":3333")
}
