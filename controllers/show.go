package controllers

import (
	"fmt"
	"log"

	"notepad/database"
	"notepad/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
)

func Show(ctx *fiber.Ctx) error {
	db := database.DbConn()

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS music(id INT, name VARCHAR(200), product VARCHAR(200), PRIMARY KEY(id))`)
	if err != nil {
		fmt.Println("cannot create table")
	}

	ins, err := db.Prepare("INSERT INTO music (name, product) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer ins.Close()

	// ret, err := ins.Exec("white strips", "elephant")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(ret)

	members := []models.Member{}
	rows, err := db.Query(`SELECT id, name, product FROM music`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		m := models.Member{}
		err = rows.Scan(&m.Id, &m.Name, &m.Product)
		if err != nil {
			log.Fatal(err)
		}
		members = append(members, m)
	}
	for i, m := range members {
		fmt.Println(i, m)
	}

	return ctx.JSON(members)
}
