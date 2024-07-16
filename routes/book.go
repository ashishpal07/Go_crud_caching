package routes

import (
	"crud_mongo/controllers"
	"github.com/gofiber/fiber/v2"
)

func AddBookGroup(app *fiber.App) {
	bookgroup := app.Group("/books")

	bookgroup.Get("/", controllers.GetBooks)
	bookgroup.Get("/:id", controllers.GetBook)
	bookgroup.Post("/", controllers.CreateBook)
	bookgroup.Delete("/:id", controllers.DeleteBook)
	bookgroup.Put("/:id", controllers.UpdateBook)
}
