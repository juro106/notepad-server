package middleware

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
)

func IsAuthenticate(ctx *fiber.Ctx) error {
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
	fmt.Println(token.UID)

	return ctx.Next()
}

func GetUserID(ctx *fiber.Ctx) (string, error) {
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

	return token.UID, nil
}
