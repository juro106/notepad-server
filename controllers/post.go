package controllers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"notepad/database"
	"notepad/middleware"
	"notepad/models"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

func makeImageDir(dir string) {
	dirname := "./images/" + dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dirname, 0777)
	}
}

type JsonObject map[string]interface{}

func (j *JsonObject) Scan(src interface{}) error {
	var _src []byte
	switch src.(type) {
	case []byte:
		_src = src.([]byte)
	default:
		return errors.New("failed to scan JsonObject")
	}
	if err := json.NewDecoder(bytes.NewReader(_src)).Decode(j); err != nil {
		return err
	}
	return nil
}

func (j JsonObject) Value() (driver.Value, error) {
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)
	if err := json.NewEncoder(buf).Encode(j); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GetContent(ctx *fiber.Ctx) error {

	db := database.DbConn()
	var query models.Query

	if err := ctx.BodyParser(&query); err != nil {
		log.Println(err)
		return err
	}
	// fmt.Println("query", query)

	var tableName string
	var defaultUser = os.Getenv("DEFAULT_USER")
	if len(query.Uid) > 0 {
		tableName = query.Uid
	} else {
		tableName = defaultUser
	}
	// table := "posts3"
	stmt := `SELECT data FROM ` + tableName + ` WHERE slug = ?`
	// fmt.Println(stmt)
	var j JsonObject
	err := db.QueryRow(stmt, query.Slug).Scan(&j)
	if err != nil {
		log.Println(err)
	}
	// stmt2 := `SELECT updated_at, data FROM ` + tableName + ` WHERE slug = ?`
	stmt2 := `SELECT json_object('updated_at', updated_at, 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags') FROM ` + tableName + ` WHERE slug = ?`
	// fmt.Println(stmt2)
	var j2 JsonObject
	err = db.QueryRow(stmt2, query.Slug).Scan(&j2)
	if err != nil {
		log.Println(err)
	}

	// fmt.Println(j)
	return ctx.JSON(j2)
}

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

func GetContentN(ctx *fiber.Ctx) error {
	msg := ctx.Params("projects") + ctx.Params("slug")

	return ctx.SendString("project & slug: " + msg)
}

func PostContent(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var content models.Content
	// fmt.Println(content)
	if err := ctx.BodyParser(&content); err != nil {
		log.Println(err)
		return err
	}
	stmt := `INSERT INTO ` + content.Project + ` (slug, data) VALUES(?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data)`
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
		stmt2 := `INSERT INTO ` + content.Project + ` (slug, data) VALUES(?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data)`
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
			return err
		}
	}

	var u string
	stmt4 := `SELECT updated_at FROM ` + content.Project + ` WHERE slug = ?`
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

func Post(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var member models.Member

	if err := ctx.BodyParser(&member); err != nil {
		return err
	}

	fmt.Println(member)

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS music(id INT, name VARCHAR(200), product VARCHAR(200), PRIMARY KEY(id))`)
	if err != nil {
		fmt.Println("cannot create table")
	}

	ins, err := db.Prepare("INSERT INTO music (name, product) VALUES(?, ?)")
	if err != nil {
		log.Println(err)
	}
	defer ins.Close()

	ret, err := ins.Exec(&member.Name, &member.Product)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(ret)

	return ctx.JSON(member)
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
	stmt := `DELETE FROM ` + tableName + ` WHERE slug = ?`
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
