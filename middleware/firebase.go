package middleware

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"github.com/gofiber/fiber/v2"
)

func SecretUserInfo(ctx *fiber.Ctx) (string, error) {
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
