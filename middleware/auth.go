package middleware

import (
	"context"
	"fmt"
	"log"

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

func SetUserInfo(c *fiber.Ctx) error {
	fmt.Println("SetUserInfo")
	uid, err := GetUID(c)
	fmt.Println("SetUserInfo -> uid", uid)

	sess, err := store.Get(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}

	sess.Set("name", uid)
	// name := sess.Get("name")
	// fmt.Println("SetUserInfo -> name", name)

	sid := sess.ID()
	if err := sess.Save(); err != nil {
		panic(err)
	}
	//
	fmt.Printf("sid %+v\n", sid)
	fmt.Printf("sess %+v\n", sess)
	log.Printf("uid: %v\n", uid)

	return c.JSON(sid)
}

func IsAuthenticate(c *fiber.Ctx) error {
	// cookie に紐付いた セッション変数の uid をチェックする
	// uid, err := GetUID(c)
	fmt.Println("use middleware")
	sess, err := store.Get(c)
	if err != nil {
		log.Fatalf("session err %v\n", err)
	}
	c.Cookies("name")
	cid := c.Cookies("cid")
	fmt.Printf("middleware cid:%+v\n", cid)

	uid := sess.Get(cid)
	fmt.Printf("middleware session uid:%+v\n", uid)
	name := sess.Get("name")
	fmt.Printf("middleware session name uid:%+v\n", name)

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
	fmt.Printf("middleware session name uid:%+v\n", name)
	nameStr := fmt.Sprintf("%s", name)

	return nameStr
}
