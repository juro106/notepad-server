package controllers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	_ "encoding/json"
	"errors"
	"fmt"
	"log"

	"notepad/database"
	"notepad/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

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
	var title models.Title

	if err := ctx.BodyParser(&title); err != nil {
		log.Println(err)
		return err
	}
	var j JsonObject
	err := db.QueryRow(`SELECT data FROM posts2 WHERE data->"$.title" = ?`, title.Title).Scan(&j)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(j)
	return ctx.JSON(j)
}

func GetContentsAll(ctx *fiber.Ctx) error {
	db := database.DbConn()
	rows, err := db.Query(`SELECT data FROM posts2 ORDER BY updated_at DESC`)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	var js []JsonObject
	for rows.Next() {
		var j JsonObject
		if err := rows.Scan(&j); err != nil {
			log.Println(err)
		}
		js = append(js, j)
	}

	fmt.Println(js)
	return ctx.JSON(js)
}

func PostContent(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var content models.Content
	// fmt.Println(content)
	if err := ctx.BodyParser(&content); err != nil {
		log.Println(err)
		return err
	}
	fmt.Println(content)
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS posts2 (
        id int NOT NULL AUTO_INCREMENT,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        data json NOT NULL,
        PRIMARY KEY (id))`)
	if err != nil {
		log.Println(err)
	}
	i, err := db.Prepare(`INSERT INTO posts2 (data) VALUES(?) ON DUPLICATE KEY UPDATE data->"$.title" = VALUES(?)`)
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

	r, err := i.Exec(j, &content.Title)
	if err != nil {
		fmt.Println("error: ", r)
		log.Println(err)
		return err
	}
	fmt.Println(r)

	return ctx.JSON(true)
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
