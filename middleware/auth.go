package middleware

import (
	"context"
	"fmt"
	"log"
	"notepad/database"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var store = session.New(session.Config{
	KeyLookup:      "cookie:cid",
	CookiePath:     "/",
	CookieSecure:   true,
	CookieHTTPOnly: true,
})

func GetUID(c *fiber.Ctx) (string, error) {
	fmt.Println("store", store)
	fmt.Println("middleware GetUID")
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	idToken := c.Request().Header.Peek("Authorization")

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client%v\n", err)
	}

	token, err := client.VerifyIDToken(context.Background(), string(idToken))
	if err != nil {
		log.Fatalf("middleware auth error verifying ID token: %v\n", err)
	}
	fmt.Println("middleware token.UID", token.UID)

	return token.UID, nil
}

// sessino にユーザーの情報を登録
func SetUserInfo(c *fiber.Ctx) error {
	fmt.Println("SetUserInfo")

	uid, err := GetUID(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}

	// session store
	sess, err := store.Get(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}

	// register uid
	fmt.Println("SetUserInfo -> uid", uid)
	sess.Set("name", uid)

	// register projects
	db := database.DbConn()
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
	sess.Set("projects", js)

	sid := sess.ID()
	if err := sess.Save(); err != nil {
		panic(err)
	}

	return c.JSON(sid)
}

func Logout(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}
	if err := sess.Destroy(); err != nil {
		panic(err)
	}
	return c.JSON(fiber.Map{"message": "logout"})
}

func IsAuthenticate(c *fiber.Ctx) error {
	// cookie に紐付いた セッション変数の uid をチェックする
	// uid, err := GetUID(c)
	// fmt.Println("use middleware")
	sess, err := store.Get(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}

	// c.Cookies("name")
	// cid := c.Cookies("cid")
	// fmt.Printf("middleware cid:%+v\n", cid)

	name := sess.Get("name")
	log.Printf("middleware IsAuthenticate session name uid:%+v\n", name)
	if name == nil {
		return c.JSON(fiber.Map{"message": "uid 認証エラー"})
	}

	// 現在のユーザーが保持しているプロジェクト一覧を取得
	// 他のユーザーのprojectsを見られないようにするチェック用。
	projects := sess.Get("projects").([]string)
	c.Locals("userProjects", projects)

	return c.Next()
}

func GetSessionUID(c *fiber.Ctx) string {
	fmt.Println("use middleware")
	sess, err := store.Get(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}
	c.Cookies("name")
	cid := c.Cookies("cid")
	fmt.Printf("middleware cid:%+v\n", cid)

	name := sess.Get("name")
	fmt.Printf("middleware GetSessionUID session name uid:%+v\n", name)
	nameStr := fmt.Sprintf("%s", name)

	return nameStr
}
