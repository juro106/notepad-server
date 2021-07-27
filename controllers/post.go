package controllers

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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
	var slug models.Slug

	if err := ctx.BodyParser(&slug); err != nil {
		log.Println(err)
		return err
	}
	var j JsonObject
	err := db.QueryRow(`SELECT data FROM posts3 WHERE slug = ?`, slug.Slug).Scan(&j)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(j)
	return ctx.JSON(j)
}

func GetContentsAll(ctx *fiber.Ctx) error {
	db := database.DbConn()
	rows, err := db.Query(`SELECT data FROM posts3 ORDER BY updated_at DESC`)
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

	// fmt.Println(js)
	return ctx.JSON(js)
}

func GetRelated(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var slug models.Slug

	if err := ctx.BodyParser(&slug); err != nil {
		log.Println(err)
	}
	fmt.Println("slug:", slug)
	// とあるタグ名(リクエストされたslug)を指定している記事を収集
	p, err := db.Prepare(`SELECT data FROM posts3 WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`)
	if err != nil {
		log.Println(err)
	}
	defer p.Close()

	// jsonPath := "tags"
	// str := `'"` + slug.Slug + `"'`
	str := `"` + slug.Slug + `"` // -上手くいった。多分 「'」が不要なのだろう
	// str := slug.Slug
	fmt.Println("str:", str)

	rows, err := p.Query(str)
	// fmt.Printf("p:%+v\n", p)

	// `SELECT data FROM posts3 WHERE JSON_CONTAINS(data, '"test"', '$.tags'`
	// rows, err := db.Query(`SELECT data FROM posts3 WHERE JSON_CONTAINS(data, '"memo"', '$.tags')`) // - 上手くいった
	// rows, err := db.Query(`SELECT data FROM posts3 WHERE JSON_CONTAINS(data, '"` + slug.Slug + `"', '$.tags')`) // - 上手くいった
	// p, err := db.Prepare(`SELECT data FROM posts3 WHERE JSON_CONTAINS(data, ?, '$.tags'`)

	if err != nil {
		log.Println(err)
	}
	// fmt.Printf("rows: %+v\n", rows)
	defer rows.Close()

	var js []JsonObject
	for rows.Next() {
		var j JsonObject
		if err := rows.Scan(&j); err != nil {
			log.Println(err)
		}
		js = append(js, j)
	}

	// 最終的に返す json
	var tagMapList []map[string][]JsonObject

	if len(js) != 0 {
		tagMap := map[string][]JsonObject{slug.Slug: js}
		tagMapList = append(tagMapList, tagMap)
		return ctx.JSON(tagMapList)
	} else { // 何も指定されていないのは普通の記事ページなので関連コンテンツを収集
		var j []uint8
		err := db.QueryRow(`SELECT data->'$.tags' FROM posts3 WHERE slug = ?`, slug.Slug).Scan(&j)
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%T\n", j)
		// fmt.Printf("%+v\n", j)
		str := string(j)
		str = strings.Replace(str, "[", "", 1)
		str = strings.Replace(str, "]", "", 1)
		str = strings.Replace(str, " ", "", -1)
		str = strings.Replace(str, "\"", "", -1)
		// fmt.Println(str)
		tags := strings.Split(str, ",")
		// fmt.Println(tags)
		for _, v := range tags {
			var jslist []JsonObject
			p, err := db.Prepare(`SELECT data FROM posts3 WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`)
			if err != nil {
				log.Println(err)
			}
			defer p.Close()
			tag := `"` + v + `"` // -上手くいった。多分 「'」が不要なのだろう
			rows, err := p.Query(tag)
			if err != nil {
				log.Println(err)
			}
			defer rows.Close()
			// tagHead := JsonObject{
			// 	"user":    "tagName",
			// 	"title":   v,
			// 	"slug":    v,
			// 	"content": "",
			// }
			// jslist = append(jslist, tagHead)
			for rows.Next() {
				var j JsonObject
				if err := rows.Scan(&j); err != nil {
					log.Println(err)
				}
				jslist = append(jslist, j)
			}
			tagMap := map[string][]JsonObject{v: jslist}
			tagMapList = append(tagMapList, tagMap)
		}
		fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
		return ctx.JSON(tagMapList)
	}
}

func GetRelatedOnly(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var tags models.Tags
	if err := ctx.BodyParser(&tags); err != nil {
		log.Println(err)
	}
	// 最終的に返す json
	var tagMapList []map[string][]JsonObject
	for _, v := range tags.Tags {
		var jslist []JsonObject
		p, err := db.Prepare(`SELECT data FROM posts3 WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`)
		if err != nil {
			log.Println(err)
		}
		defer p.Close()
		tag := `"` + v + `"` // -上手くいった。多分 「'」が不要なのだろう
		rows, err := p.Query(tag)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		// tagHead := JsonObject{
		// 	"user":    "tagName",
		// 	"title":   v,
		// 	"slug":    v,
		// 	"content": "",
		// }
		// jslist = append(jslist, tagHead)
		for rows.Next() {
			var j JsonObject
			if err := rows.Scan(&j); err != nil {
				log.Println(err)
			}
			jslist = append(jslist, j)
		}
		tagMap := map[string][]JsonObject{v: jslist}
		tagMapList = append(tagMapList, tagMap)
	}

	fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
	return ctx.JSON(tagMapList)
}

func PostContent(ctx *fiber.Ctx) error {
	db := database.DbConn()
	var content models.Content
	// fmt.Println(content)
	if err := ctx.BodyParser(&content); err != nil {
		log.Println(err)
		return err
	}
	// fmt.Println(content)
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS posts3 (
        id int NOT NULL AUTO_INCREMENT,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        slug varchar(200) NOT NULL UNIQUE,
        data json,
        PRIMARY KEY (id))`)
	if err != nil {
		log.Println(err)
	}
	i, err := db.Prepare(`INSERT INTO posts3 (slug, data) VALUES(?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data)`)
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
		i, err := db.Prepare(`INSERT INTO posts3 (slug, data) VALUES(?, ?) ON DUPLICATE KEY UPDATE data = VALUES(data)`)
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
	err = db.QueryRow(`SELECT updated_at FROM posts3 WHERE slug = ?`, content.Slug).Scan(&u)
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
	var slug models.Slug

	if err := ctx.BodyParser(&slug); err != nil {
		log.Println(err)
		return err
	}
	p, err := db.Prepare(`DELETE FROM posts3 WHERE slug = ?`)
	if err != nil {
		log.Println(err)
	}
	defer p.Close()

	p.Exec(slug.Slug)
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
