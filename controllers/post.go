package controllers

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"context"
	"notepad/database"
	"notepad/middleware"
	"notepad/models"

	firebase "firebase.google.com/go"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	_ "github.com/go-sql-driver/mysql"
)

func makeImageDir(dir string) {
	dirname := "./images/" + dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dirname, 777)
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

	// 画像ディレクトリ
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
	// var tableName string
	// var defaultUser = os.Getenv("DEFAULT_USER")
	// if len(query.Uid) > 0 {
	// 	tableName = query.Project
	// } else {
	// 	tableName = defaultUser
	// }
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

func NewSessionID() string {
	// session ID 発行
	b := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DeleteCookie(ctx *fiber.Ctx, arg string) {
	ctx.Cookie(&fiber.Cookie{
		Name: arg,
		// Expires: time.Now().Add(-(time.Hour * 2)),
		// Expires: time.Now().Add(24 * time.Hour),
		Expires: time.Now().Add(-3 * time.Second),
	})
	ctx.ClearCookie(arg)
}

var store = session.New(session.Config{
	KeyLookup:      "cookie:cid",
	CookiePath:     "/",
	CookieSecure:   true,
	CookieHTTPOnly: true,
})

func SecretUserInfo(ctx *fiber.Ctx) error {
	fmt.Println("SecretUserInfo")
	// DeleteCookie(ctx, "token")
	// DeleteCookie(ctx, "John")
	// DeleteCookie(ctx, "sesseion_id")
	// ctx.ClearCookie("token")
	// ctx.ClearCookie("session_id")
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	idToken := ctx.Request().Header.Peek("Authorization")

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client%v\n", err)
	}

	token, err := client.VerifyIDToken(context.Background(), string(idToken))
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
	}

	// sessionID := NewSessionID()
	//
	// // cookie
	// cookie := fiber.Cookie{
	// 	Name:  "session_id",
	// 	Value: sessionID,
	// 	// Name:     "John",
	// 	// Value:    "",
	// 	Secure:   true,
	// 	Path:     "/",
	// 	HTTPOnly: true,
	// 	// Expires:  time.Now().Add(24 * time.Hour),
	// }
	// ctx.Cookie(&cookie)
	// fmt.Printf("cookie %+v\n", cookie)
	// ctx.Cookie(&fiber.Cookie{
	// 	Name:     "token",
	// 	Value:    "randomValue",
	// 	Expires:  time.Now().Add(24 * time.Hour),
	// 	HTTPOnly: true,
	// })
	// // session store
	// fmt.Printf("%T\n", store)
	// // set cookie
	//
	sess, err := store.Get(ctx)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}
	sess.Set("name", token.UID)
	name := sess.Get("name")
	fmt.Println("name", name)

	if err := sess.Save(); err != nil {
		panic(err)
	}
	//
	sid := sess.ID()
	fmt.Printf("sid %+v\n", sid)
	fmt.Printf("sess %+v\n", sess)
	// fmt.Printf("sid: %+v\n", sid)

	// log.Printf("store: %+v\n", store)

	log.Printf("token: %v\n", token.UID)

	return ctx.JSON(token)
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
