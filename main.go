package main

import (
	"crud_mongo/config"
	"crud_mongo/routes"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	err := run()

	if err != nil {
		panic(err)
	}
}

func run() error {
	err := config.LoadEnv()

	if err != nil {
		return err
	}

	err = config.InitDB()

	if err != nil {
		return err
	}

	defer config.CloseDB()

	err = config.ConnectRedis()

	if err != nil {
		return err
	}

	// defer config.CloseRedisDB()

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	routes.AddBookGroup(app)

	var port string

	port = os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	app.Listen(":" + port)

	return nil
}
