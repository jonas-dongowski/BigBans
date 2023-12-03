package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"nmbr.one/big-bans/mojangapi"
)

type Ban struct {
	gorm.Model
	BannedUUID string
	BannerUUID string
	Duration   time.Duration
}

type Connection struct {
	gorm.Model
	PlayerUUID   string
	LastActiveAt time.Time
	Token        string
}

type IndexView struct {
	Token string
}

type BansView struct {
	Bans []BansViewBan
}

type BansViewBan struct {
	Ban
	CreatedAtF string
	BannedName string
	BannerName string
}

type FrameView struct {
	PlayerUUID string
}

func main() {
	dsn := "root:localroot@tcp(127.0.0.1:3999)/bigbans?charset=utf8mb4&parseTime=True&loc=Local"
	db, dberr := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			NoLowerCase:   true,
		},
	})

	if dberr != nil {
		panic("Failed to connect to local database")
	}

	db.AutoMigrate(&Ban{})
	db.AutoMigrate(&Connection{})

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		token := c.Query("token")

		return c.Render("index", IndexView{
			Token: token,
		})
	})

	app.Get("/bans", func(c *fiber.Ctx) error {
		var token = c.Query("token")
		var connections []Connection

		db.Where("Token = ?", token).Find(&connections)

		if len(connections) > 0 {
			var con = connections[0]

			if !time.Now().After(con.LastActiveAt.Add(time.Hour * 1)) {
				con.LastActiveAt = time.Now()

				db.Save(con)

				// Session good
				var bans []Ban
				var viewBans []BansViewBan

				db.Find(&bans)

				for _, ban := range bans {
					bannedName, err := mojangapi.GetNameFromUUID(ban.BannedUUID)

					if err != nil {
						bannedName = "ERROR: " + err.Error()
					}

					bannerName, err := mojangapi.GetNameFromUUID(ban.BannerUUID)

					if err != nil {
						bannerName = "ERROR: " + err.Error()
					}

					var viewBan = BansViewBan{
						Ban:        ban,
						CreatedAtF: ban.CreatedAt.Format(time.RFC1123),
						BannedName: bannedName,
						BannerName: bannerName,
					}

					viewBans = append(viewBans, viewBan)
				}

				return c.Render("bans", BansView{
					Bans: viewBans,
				})
			}
		}

		return c.Render("invalid_session", nil)
	})

	app.Post("/logout", func(c *fiber.Ctx) error {
		var body = struct {
			Token string `json:"token"`
		}{}
		var bodyerror = c.BodyParser(&body)

		if bodyerror != nil {
			return c.SendStatus(400)
		}

		var connections []Connection

		db.Where("Token = ?", body.Token).Find(&connections)

		if len(connections) > 0 {
			var con = connections[0]

			if !time.Now().After(con.LastActiveAt.Add(time.Hour * 1)) {
				db.Delete(&con)

				return c.Render("logged_out", nil)
			}
		}

		return c.SendStatus(401)
	})

	log.Fatal(app.Listen(":3000"))
}
