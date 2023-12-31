package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ztrue/tracerr"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	WebServerPort    int    `json:"web_server_port"`
	ExternalURL      string `json:"external_url"`
	PreSharedKey     string `json:"psk"`
	DatabaseFilePath string `json:"db_file_path"`
}

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

func configExists() bool {
	_, err := os.Stat("../config.json")

	return err == nil
}

func getConfig() (*Config, error) {
	config := new(Config)
	data, fileError := os.ReadFile("../config.json")

	if fileError != nil {
		return nil, fileError
	}

	jsonError := json.Unmarshal(data, config)

	if jsonError != nil {
		return nil, jsonError
	}

	return config, nil
}

func createLogFile() {
	file, _ := os.Create("./log.txt")

	file.Close()
}

func appendToLogFile(msg string) {
	file, _ := os.OpenFile("./log.txt", os.O_APPEND|os.O_WRONLY, os.ModePerm)

	file.WriteString("[" + time.Now().Format(time.RFC1123) + "] " + msg)
	file.WriteString("\n")

	file.Close()
}

func main() {
	// dsn := "root:localroot@tcp(127.0.0.1:3999)/bigbans?charset=utf8mb4&parseTime=True&loc=Local"
	// db, dberr := gorm.Open(mysql.Open(dsn), &gorm.Config{
	// 	NamingStrategy: schema.NamingStrategy{
	// 		SingularTable: false,
	// 		NoLowerCase:   true,
	// 	},
	// })
	createLogFile()

	var config *Config

	if configExists() {
		var newConfig, configError = getConfig()

		if configError != nil {
			appendToLogFile(tracerr.SprintSource(configError))
			return
		}

		config = newConfig
	} else {
		appendToLogFile(tracerr.SprintSource(errors.New("Config does not exist")))
		return
	}

	db, dberr := gorm.Open(sqlite.Open(config.DatabaseFilePath), &gorm.Config{})

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
		var psk = c.Get("X-PSK")

		if psk != config.PreSharedKey {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		var body = new(CreateSession)
		var bodyError = c.BodyParser(&body)

		if bodyError != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		var session = Session{
			PlayerUUID: body.PlayerUUID,
			Token:      generateRandomString(16),
		}

		db.Save(&session)

		var createdSession = CreatedSession{
			Token: session.Token,
		}

		return c.JSON(createdSession)
	})

	log.Fatal(app.Listen(":" + strconv.Itoa(config.WebServerPort)))
}
