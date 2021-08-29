package controllers

import (
	"fmt"
	"log"
	"notepad/database"
	"notepad/middleware"
	"os"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

func DeleteImage(c *fiber.Ctx) error {
	dirName := c.Params("project")
	fileName := c.Params("filename")
	path := "./images/" + dirName + "/" + fileName

	log.Println("delete", path)

	if err := os.Remove(path); err != nil {
		log.Println("delete err:", err)
	}

	return c.JSON(fiber.Map{"status": 201, "message": "success", "data": path})
}

func DeleteProject(c *fiber.Ctx) error {
	projectName := c.Params("project")
	db := database.DbConn()

	fmt.Println(projectName)
	// ユーザーとプロジェクトを紐付けるテーブルから項目を削除
	stmt := `DELETE FROM projects WHERE name = ?`
	p, err := db.Prepare(stmt)
	if err != nil {
		log.Println(err)
	}
	defer p.Close()

	p.Exec(projectName)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("delete ", projectName, " from projectsTable")
	// テーブルを削除
	stmt2 := `DROP TABLE IF EXISTS ` + projectName
	p2, err := db.Prepare(stmt2)
	if err != nil {
		log.Println(err)
	}
	defer p2.Close()

	p2.Exec(projectName)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("delete table:", projectName)

	// 画像ディレクトリを削除
	if err := os.Remove("./images/" + projectName); err != nil {
		log.Println(err)
	}
	log.Println("delete images directory:", projectName)

	middleware.DeleteProjectInTheSession(c, projectName)

	return c.JSON(fiber.Map{"status": 201, "message": "success", "data": projectName})
}
