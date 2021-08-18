package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notepad/database"
	"notepad/middleware"
	"notepad/models"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

// method: post

func CreateTable(c *fiber.Ctx) error {
	db := database.DbConn()
	type Name struct {
		Name string `json:"name"`
	}
	var name Name
	if err := c.BodyParser(&name); err != nil {
		log.Println(err)
	}

	stmt := `CREATE TABLE ` + name.Name + ` (
        slug varchar(200) NOT NULL UNIQUE,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        data json,
        PRIMARY KEY (slug))`
	// fmt.Println(content)

	_, err := db.Exec(stmt)

	// // 認証情報に新しいテーブルを追加
	// middleware.IsAuthenticate(c)

	// 画像ディレクトリも一緒に作成
	makeImageDir(name.Name)

	ins, err := db.Prepare("INSERT INTO projects (name, owner) VALUES(?, ?)")
	if err != nil {
		log.Println(err)
	}
	defer ins.Close()
	uid := middleware.GetSessionUID(c)
	ret, err := ins.Exec(name.Name, uid)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(ret)

	var message models.Message
	t := fmt.Sprintf("%s", time.Now())
	message.UpdatedAt = t
	if err != nil {
		log.Println(err)
		message.Message = fmt.Sprintf("%s", err)
		return c.JSON(message)
	} else {
		message.Message = "success"
		return c.JSON(message)
	}
}

func PostContent(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var content models.Content
	// fmt.Println(content)
	if err := ctx.BodyParser(&content); err != nil {
		log.Println(err)
		return err
	}
	stmt := "INSERT INTO `" + content.Project + "` (slug, data) VALUES(?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data)"
	i, err := db.Prepare(stmt)
	if err != nil {
		fmt.Println("error: ", i)
		log.Println(err)
		return err
	}
	defer i.Close()

	j, err := json.Marshal(&content)
	if err != nil {
		log.Println(err)
	}

	r, err := i.Exec(&content.Slug, j)
	if err != nil {
		fmt.Println("error: ", r)
		log.Println(err)
		return err
	}
	// fmt.Println(r)

	for _, v := range content.Tags {
		stmt2 := "INSERT INTO `" + content.Project + "` (slug, data) VALUES(?, ?)"
		i, err := db.Prepare(stmt2)
		if err != nil {
			log.Print(err)
		}
		defer i.Close()
		t := models.Content{
			User:    content.User,
			Title:   v,
			Slug:    v,
			Tags:    []string{},
			Content: "",
		}
		j, err := json.Marshal(&t)
		if err != nil {
			log.Println(err)
		}
		r, err := i.Exec(v, j)
		if err != nil {
			fmt.Println("error: ", r)
			log.Println(err)
		}
	}

	var u string
	stmt4 := "SELECT updated_at FROM `" + content.Project + "` WHERE slug = ?"
	err = db.QueryRow(stmt4, content.Slug).Scan(&u)
	if err != nil {
		log.Println(err)
	}
	message := models.Message{
		UpdatedAt: u,
		Message:   "success",
	}
	return ctx.JSON(message)
}

func DeleteContent(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var query models.Query
	if err := ctx.BodyParser(&query); err != nil {
		log.Println(err)
		return err
	}
	tableName := query.Project
	fmt.Println(query)
	stmt := "DELETE FROM `" + tableName + "` WHERE slug = ?"
	p, err := db.Prepare(stmt)
	if err != nil {
		log.Println(err)
	}
	defer p.Close()

	p.Exec(query.Slug)
	if err != nil {
		log.Println(err)
		return err
	} else {
		t := fmt.Sprintf("%s", time.Now())
		message := models.Message{
			UpdatedAt: t,
			Message:   "success",
		}
		return ctx.JSON(message)
	}
}

func UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("image upload error --->", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}
	fmt.Printf("image filename:%+v\n", file.Filename)

	projectName := c.FormValue("project")
	if len(projectName) == 0 {
		log.Println("get project name error --->")
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}
	fmt.Printf("projectName:%+v\n", projectName)

	filename := "./images/" + projectName + "/" + file.Filename

	err = c.SaveFile(file, filename)
	if err != nil {
		log.Println("image save error --->", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	data := map[string]interface{}{
		"imageName":   filename,
		"projectName": projectName,
		"header":      file.Header,
		"size":        file.Size,
	}

	return c.JSON(fiber.Map{"status": 201, "messeage": "success", "data": data})
}
