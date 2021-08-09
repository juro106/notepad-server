package controllers

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

func DeleteImage(c *fiber.Ctx) error {
	dirName := c.Params("projects")
	fileName := c.Params("filename")
	path := "./images/" + dirName + "/" + fileName

	log.Println("delete", path)

	if err := os.Remove(path); err != nil {
		log.Println("delete err:", err)
	}

	return c.JSON(fiber.Map{"status": 201, "message": "success", "data": path})
}
