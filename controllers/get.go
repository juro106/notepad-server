package controllers

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"

	"notepad/database"
	"notepad/middleware"
	"notepad/models"

	"github.com/gofiber/fiber/v2"

	_ "github.com/go-sql-driver/mysql"
)

// method: get

func GetContents(c *fiber.Ctx) error {

	db := database.DbConn()

	tableName := c.Params("project")
	slug := c.Params("slug")

	tableName, err := url.QueryUnescape(tableName)
	if err != nil {
		log.Println(err)
	}
	slug, err = url.QueryUnescape(slug)
	if err != nil {
		log.Println(err)
	}
	// 他のユーザーのプロジェクトは閲覧できないようにする
	authProject := checkUserProjects(tableName, c.Locals("userProjects").([]string))
	if !authProject {
		return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	}

	stmt := `SELECT json_object('updated_at', updated_at, 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags', 'image', data->'$.image') FROM ` + tableName + ` WHERE slug = ?`
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
	tableName := c.Params("project")
	tableName, err := url.QueryUnescape(tableName)
	if err != nil {
		log.Println(err)
	}
	// 他のユーザーのプロジェクトは閲覧できないようにする
	authProject := checkUserProjects(tableName, c.Locals("userProjects").([]string))
	if !authProject {
		return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	}

	stmt := "SELECT json_object('updated_at', date_format(updated_at, '%Y-%m-%d'), 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags', 'image', data->'$.image') FROM `" + tableName + "` ORDER BY updated_at DESC"
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
		log.Println("PreparedStatement error", err)
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
	// fmt.Printf("%v\n", js)
	return c.JSON(js)
}

func GetImages(c *fiber.Ctx) error {
	dirName := c.Params("project")
	dirPath := "./images/" + dirName
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	type FileData struct {
		Name    string `json:"name"`
		ModTime string `json:"modTime"`
	}
	flist := []FileData{}
	for _, f := range files {
		// fmt.Printf("%+v", f)
		filename := "/images/" + dirName + "/" + f.Name()
		finfo, err := f.Info()
		if err != nil {
			log.Fatal(err)
		}
		modTime := finfo.ModTime()
		filedata := FileData{
			Name:    filename,
			ModTime: modTime.Format("2006-01-02 15:04:05"),
		}

		flist = append(flist, filedata)
	}
	// 日付順にソート
	sort.Slice(flist, func(i, j int) bool {
		return time2int(flist[i].ModTime) > time2int(flist[j].ModTime)
	})
	// fmt.Printf("flist: %+v\n", flist)

	return c.JSON(flist)
}

func GetRelated(c *fiber.Ctx) error {
	db := database.DbConn()

	tableName := c.Params("project")
	slug := c.Params("slug")
	tableName, err := url.QueryUnescape(tableName)
	if err != nil {
		log.Println(err)
	}
	slug, err = url.QueryUnescape(slug)
	if err != nil {
		log.Println(err)
	}
	// 他のユーザーのプロジェクトは閲覧できないようにする
	authProject := checkUserProjects(tableName, c.Locals("userProjects").([]string))
	if !authProject {
		return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	}

	// var defaultUser = os.Getenv("DEFAULT_USER") // ※public用の処理で使う予定

	// とあるタグ名(リクエストされたslug)を指定している記事を収集
	stmt := "SELECT data FROM `" + tableName + "` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')"
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

	var js []ContentObject
	for rows.Next() {
		var j ContentObject
		if err := rows.Scan(&j); err != nil {
			log.Println(err)
		}
		js = append(js, j)
	}

	// 最終的に返す json
	var tagMapList []map[string][]ContentObject

	if len(js) != 0 {
		tagMap := map[string][]ContentObject{slug: js}
		tagMapList = append(tagMapList, tagMap)
		return c.JSON(tagMapList)
	} else { // 何も指定されていないのは普通の記事ページなので関連コンテンツを収集
		var t TagsObject
		stmt := "SELECT JSON_OBJECT('tags', data->'$.tags') FROM `" + tableName + "` WHERE slug = ?"
		err := db.QueryRow(stmt, slug).Scan(&t)
		if err != nil {
			log.Println(err)
		}
		// fmt.Println(tags)
		for _, v := range t.Tags {
			var jslist []ContentObject
			stmt := "SELECT data FROM `" + tableName + "` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags') ORDER BY updated_at DESC"
			// SELECT * FROM public WHERE JSON_CONTAINS(data, '"image"', '$.tags') ORDER BY updated_at DESC
			p, err := db.Prepare(stmt)
			if err != nil {
				log.Println(err)
			}
			defer p.Close()
			tag := `"` + v + `"`
			rows, err := p.Query(tag)
			if err != nil {
				log.Println(err)
			}
			defer rows.Close()

			for rows.Next() {
				var j ContentObject
				if err := rows.Scan(&j); err != nil {
					log.Println(err)
				}
				jslist = append(jslist, j)
			}
			tagMap := map[string][]ContentObject{v: jslist}
			tagMapList = append(tagMapList, tagMap)
		}
		// fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
		return c.JSON(tagMapList)
	}
}

func GetTags(c *fiber.Ctx) error {
	db := database.DbConn()
	tableName := c.Params("project")
	log.Println("query check", tableName)
	// 他のユーザーのプロジェクトは閲覧できないようにする
	authProject := checkUserProjects(tableName, c.Locals("userProjects").([]string))
	if !authProject {
		return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	}

	stmt := "SELECT DISTINCT JSON_OBJECT('tags', data->'$.tags') FROM `" + tableName + "`"
	// stmt = `SELECT DISTINCT JSON_EXTRACT(data, '$.tags') FROM public`
	p, err := db.Prepare(stmt)
	if err != nil {
		log.Println("tags get", err)
	}
	defer p.Close()

	var js []TagsObject
	// var js []JsonObject
	rows, err := p.Query()
	if err != nil {
		log.Println("p.Query", err)
		return c.JSON(js)
	}
	defer rows.Close()

	for rows.Next() {
		var j TagsObject
		if err := rows.Scan(&j); err != nil {
			log.Println("json scan", err)
		}
		js = append(js, j)
	}

	tmpTagList := []string{}
	for _, m := range js {
		tags := m.Tags
		for _, tag := range tags {
			tmpTagList = append(tmpTagList, tag)
		}
	}
	l := len(tmpTagList)
	tagMap := make(map[string]struct{}, l)
	tagList := make([]string, 0, l)

	for _, elem := range tmpTagList {
		// mapの第2引数には、その値が入っているかどうかの真偽値が入っている。
		if _, ok := tagMap[elem]; !ok && len(elem) != 0 {
			tagMap[elem] = struct{}{}
			tagList = append(tagList, elem)
		}
	}

	sort.Strings(tagList)
	// log.Println(tagList)

	var tagNumList []TagNumObject
	for _, v := range tagList {
		stmt = "SELECT COUNT(*) FROM `" + tableName + "` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')"

		p, err := db.Prepare(stmt)
		if err != nil {
			log.Println(err)
		}
		defer p.Close()
		tag := `"` + v + `"` // -上手くいった。多分 「'」が不要なのだろう

		// var num にsqlで集計した「数」を格納
		var num int
		if err := p.QueryRow(tag).Scan(&num); err != nil {
			log.Println(err)
		}
		t := TagNumObject{
			Name:   v,
			Number: num,
		}
		tagNumList = append(tagNumList, t)
	}

	return c.JSON(tagNumList)
}

// これはPOST
func GetRelatedOnly(c *fiber.Ctx) error {
	db := database.DbConn()
	var query models.Query

	if err := c.BodyParser(&query); err != nil {
		log.Println(err)
	}
	// fmt.Printf("query%+v\n", query)

	// 他のユーザーのプロジェクトは閲覧できないようにする
	authProject := checkUserProjects(query.Project, c.Locals("userProjects").([]string))
	if !authProject {
		return c.Status(fiber.StatusNotFound).SendString("Sorry can't find that!")
	}

	var tagMapList []map[string][]ContentObject // 最終的に返す json
	for _, v := range query.Tags {
		var jslist []ContentObject
		stmt := "SELECT json_object('updated_at', DATE_FORMAT(updated_at, '%Y%m%d%T'), 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags', 'image', data->'$.image') FROM `" + query.Project + "` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')"
		// stmt := `SELECT data, updated_at FROM ` + query.Project + ` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')`
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
			var j ContentObject
			if err := rows.Scan(&j); err != nil {
				log.Println(err)
			}
			fmt.Println(time2int(j.Updated_at))
			jslist = append(jslist, j)
		}
		// 日付順にソート
		sort.Slice(jslist, func(i, j int) bool {
			return time2int(jslist[i].Updated_at) > time2int(jslist[j].Updated_at)
		})
		// fmt.Printf("%+v\n", jslist)

		tagMap := map[string][]ContentObject{v: jslist}
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

	slug, err := url.QueryUnescape(slug)
	if err != nil {
		log.Println(err)
	}

	stmt := "SELECT json_object('updated_at', updated_at, 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags', 'image', data->'$.image') FROM `" + tableName + "` WHERE slug = ?"
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

	stmt := "SELECT json_object('updated_at', date_format(updated_at, '%Y-%m-%d'), 'slug', data->'$.slug', 'user', data->'$.user', 'content', data->'$.content', 'title', data->'$.title', 'tags', data->'$.tags', 'image', data->'$.image') FROM `" + tableName + "` ORDER BY updated_at DESC"
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

	slug, err := url.QueryUnescape(slug)
	if err != nil {
		log.Println(err)
	}
	// var defaultUser = os.Getenv("DEFAULT_USER") // ※public用の処理で使う予定

	// とあるタグ名(リクエストされたslug)を指定している記事を収集
	stmt := "SELECT data FROM `" + tableName + "` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')"
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

	var js []ContentObject
	for rows.Next() {
		var j ContentObject
		if err := rows.Scan(&j); err != nil {
			log.Println(err)
		}
		js = append(js, j)
	}

	// 最終的に返す json
	var tagMapList []map[string][]ContentObject

	if len(js) != 0 {
		tagMap := map[string][]ContentObject{slug: js}
		tagMapList = append(tagMapList, tagMap)
		return c.JSON(tagMapList)
	} else { // 何も指定されていないのは普通の記事ページなので関連コンテンツを収集
		var t TagsObject
		stmt := "SELECT JSON_OBJECT('tags', data->'$.tags') FROM `" + tableName + "` WHERE slug = ?"
		err := db.QueryRow(stmt, slug).Scan(&t)
		if err != nil {
			log.Println(err)
		}
		// fmt.Println(tags)
		for _, v := range t.Tags {
			var jslist []ContentObject
			stmt := "SELECT data FROM `" + tableName + "` WHERE JSON_CONTAINS(data, CAST(? AS JSON), '$.tags')"
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
				var j ContentObject
				if err := rows.Scan(&j); err != nil {
					log.Println(err)
				}
				jslist = append(jslist, j)
			}
			tagMap := map[string][]ContentObject{v: jslist}
			tagMapList = append(tagMapList, tagMap)
		}
		// fmt.Printf("######\n\ntagMapList: %+v\n", tagMapList)
		return c.JSON(tagMapList)
	}
}
