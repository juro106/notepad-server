package controllers

import (
	"fmt"
	"log"
	"os"
	"strings"

	"notepad/database"
	"notepad/middleware"
	"notepad/models"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

func GetContents(c *fiber.Ctx) error {

	db := database.DbConn()

	tableName := c.Params("projects")
	slug := c.Params("slug")

	stmt := `SELECT json_object('updated_at', updated_at, 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags', 'image', data->'$.image') FROM ` + tableName + ` WHERE slug = ?`
	// fmt.Println(stmt2)
	var j JsonObject
	if err := db.QueryRow(stmt, slug).Scan(&j); err != nil {
		log.Println(err)
	}

	fmt.Println(j)
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

func GetImages(c *fiber.Ctx) error {
	dirName := c.Params("project")
	dirPath := "./images/" + dirName
	f, err := os.Open(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	files, _ := f.Readdirnames(0)
	flist := []string{}
	for _, f := range files {
		filename := "/images/" + dirName + "/" + f
		flist = append(flist, filename)
	}
	fmt.Printf("flist: %+v", flist)

	return c.JSON(flist)
}

func GetRelated(c *fiber.Ctx) error {
	db := database.DbConn()

	tableName := c.Params("projects")
	slug := c.Params("slug")

	// var defaultUser = os.Getenv("DEFAULT_USER") // ※public用の処理で使う予定

	// とあるタグ名(リクエストされたslug)を指定している記事を収集
	stmt := `SELECT data FROM ` + tableName + ` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`
	p, err := db.Prepare(stmt)
	if err != nil {
		log.Println(err)
	}
	defer p.Close()

	str := `"` + slug + `"` // -上手くいった。多分 「'」が不要なのだろう
	rows, err := p.Query(str)
	// fmt.Printf("p:%+v\n", p)

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
		tagMap := map[string][]JsonObject{slug: js}
		tagMapList = append(tagMapList, tagMap)
		return c.JSON(tagMapList)
	} else { // 何も指定されていないのは普通の記事ページなので関連コンテンツを収集
		var j []uint8
		stmt := `SELECT data->'$.tags' FROM ` + tableName + ` WHERE slug = ?`
		err := db.QueryRow(stmt, slug).Scan(&j)
		if err != nil {
			log.Println(err)
		}
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
			stmt := `SELECT data FROM ` + tableName + ` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`
			p, err := db.Prepare(stmt)
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
		// fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
		return c.JSON(tagMapList)
	}
}

// これはPOST
func GetRelatedOnly(c *fiber.Ctx) error {
	db := database.DbConn()
	var query models.Query

	if err := c.BodyParser(&query); err != nil {
		log.Println(err)
	}
	fmt.Printf("query%+v\n", query)

	var tagMapList []map[string][]JsonObject // 最終的に返す json
	for _, v := range query.Tags {
		var jslist []JsonObject
		stmt := `SELECT data FROM ` + query.Project + ` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`
		p, err := db.Prepare(stmt)
		if err != nil {
			log.Println(err)
			return c.JSON(tagMapList)
		}
		defer p.Close()

		tag := `"` + v + `"` // -上手くいった。多分 「'」が不要なのだろう
		rows, err := p.Query(tag)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

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

	// fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
	return c.JSON(tagMapList)
}

// public
func GetPublicContents(c *fiber.Ctx) error {
	db := database.DbConn()

	tableName := os.Getenv("DEFAULT_TABLE")
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

func GetPublicContentsAll(c *fiber.Ctx) error {
	fmt.Println("GetPublicContentsAll")
	db := database.DbConn()

	tableName := os.Getenv("DEFAULT_TABLE")

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

func GetPublicRelated(c *fiber.Ctx) error {
	db := database.DbConn()

	tableName := os.Getenv("DEFAULT_TABLE")
	slug := c.Params("slug")

	// var defaultUser = os.Getenv("DEFAULT_USER") // ※public用の処理で使う予定

	// とあるタグ名(リクエストされたslug)を指定している記事を収集
	stmt := `SELECT data FROM ` + tableName + ` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`
	p, err := db.Prepare(stmt)
	if err != nil {
		log.Println(err)
	}
	defer p.Close()

	str := `"` + slug + `"` // -上手くいった。多分 「'」が不要なのだろう
	rows, err := p.Query(str)
	// fmt.Printf("p:%+v\n", p)

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
		tagMap := map[string][]JsonObject{slug: js}
		tagMapList = append(tagMapList, tagMap)
		return c.JSON(tagMapList)
	} else { // 何も指定されていないのは普通の記事ページなので関連コンテンツを収集
		var j []uint8
		stmt := `SELECT data->'$.tags' FROM ` + tableName + ` WHERE slug = ?`
		err := db.QueryRow(stmt, slug).Scan(&j)
		if err != nil {
			log.Println(err)
		}
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
			stmt := `SELECT data FROM ` + tableName + ` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`
			p, err := db.Prepare(stmt)
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
		// fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
		return c.JSON(tagMapList)
	}
}
