package controllers

import (
	"fmt"
	"log"

	"notepad/database"
	"notepad/middleware"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

func GetContents(c *fiber.Ctx) error {

	db := database.DbConn()

	tableName := c.Params("projects")
	slug := c.Params("slug")

	stmt := `SELECT json_object('updated_at', updated_at, 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags') FROM ` + tableName + ` WHERE slug = ?`
	// fmt.Println(stmt2)
	var j JsonObject
	if err := db.QueryRow(stmt, slug).Scan(&j); err != nil {
		log.Println(err)
	}

	// fmt.Println(j)
	return c.JSON(j)
}

func GetContentsAll(c *fiber.Ctx) error {
	db := database.DbConn()

	tableName := c.Params("projects")

	stmt := `SELECT json_object('updated_at', date_format(updated_at, '%Y-%m-%d'), 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title') FROM ` + tableName + ` ORDER BY updated_at DESC`
	rows, err := db.Query(stmt)
	var js []JsonObject
	if err != nil {
		log.Println(err)
		return c.JSON(js)
	}
	defer rows.Close()

	for rows.Next() {
		var j JsonObject
		if err := rows.Scan(&j); err != nil {
			log.Println(err)
		}
		js = append(js, j)
	}

	// fmt.Println(js)
	return c.JSON(js)
}

func GetProjects(c *fiber.Ctx) error {
	db := database.DbConn()
	uid := middleware.GetSessionUID(c)

	stmt := `SELECT name FROM projects WHERE owner = ?`

	p, err := db.Prepare(stmt)
	if err != nil {
		log.Println("preparedstatement error", err)
	}
	defer p.Close()

	rows, err := p.Query(uid)
	var js []string
	if err != nil {
		log.Println("rows error", err)
		return c.JSON(js)
	}
	defer rows.Close()

	for rows.Next() {
		var j string
		if err := rows.Scan(&j); err != nil {
			log.Println(err)
		}
		js = append(js, j)
	}
	fmt.Printf("%v\n", js)
	return c.JSON(js)
}
