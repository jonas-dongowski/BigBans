package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Ban struct {
	gorm.Model
	BannedUUID string
	BannerUUID string
	Duration   time.Duration
}

type Session struct {
	gorm.Model
	PlayerUUID   string
	LastActiveAt time.Time
	Token        string
}

type CreateSession struct {
	PlayerUUID string `json:"playerUniqueId"`
}

type CreatedSession struct {
	Token string `json:"token"`
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func main() {
	// dsn := "root:localroot@tcp(127.0.0.1:3999)/bigbans?charset=utf8mb4&parseTime=True&loc=Local"
	// db, dberr := gorm.Open(mysql.Open(dsn), &gorm.Config{
	// 	NamingStrategy: schema.NamingStrategy{
	// 		SingularTable: false,
	// 		NoLowerCase:   true,
	// 	},
	// })

	db, dberr := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	if dberr != nil {
		panic("Failed to connect to local database")
	}

	db.AutoMigrate(&Ban{})
	db.AutoMigrate(&Session{})

	app := fiber.New(fiber.Config{})

	app.Use(logger.New())
	app.Use(cors.New())

	app.Static("/", "public")

	app.Post("/api/sessions", func(c *fiber.Ctx) error {
		var body = new(CreateSession)
		var bodyError = c.BodyParser(&body)

		if bodyError != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		var session = Session{
			PlayerUUID: body.PlayerUUID,
			Token:      generateRandomString(32),
		}

		db.Save(&session)

		var createdSession = CreatedSession{
			Token: session.Token,
		}

		return c.JSON(createdSession)
	})

	log.Fatal(app.Listen(":3000"))
}
