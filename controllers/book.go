package controllers

import (
	"crud_mongo/config"
	dtos "crud_mongo/dtos"
	"crud_mongo/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetBooks(ctx *fiber.Ctx) error {
	collection := config.GetDBCollection("books")

	books := make([]models.Book, 0)
	getAllBook, err := collection.Find(ctx.Context(), bson.M{})

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	for getAllBook.Next(ctx.Context()) {
		book := models.Book{}
		err := getAllBook.Decode(&book)

		if err != nil {
			return ctx.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		books = append(books, book)
	}

	return ctx.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    books,
	})
}

func GetBook(ctx *fiber.Ctx) error {
	collection := config.GetDBCollection("books")

	id := ctx.Params("id")

	if id == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "id is required.",
		})
	}

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"seccess": true,
			"message": "id is invalid",
		})
	}

	book := models.Book{}

	err = collection.FindOne(ctx.Context(), bson.M{"_id": objectId}).Decode(&book)

	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"success": true,
			"message": "book not found",
		})
	}

	return ctx.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    book,
	})
}

func CreateBook(ctx *fiber.Ctx) error {
	b := new(dtos.CreateBookDto)
	err := ctx.BodyParser(&b)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid body struct.",
		})
	}

	collection := config.GetDBCollection("books")
	result, err := collection.InsertOne(ctx.Context(), b)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "failed to create book",
			"error":   err.Error(),
		})
	}

	return ctx.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

func UpdateBook(ctx *fiber.Ctx) error {
	b := new(dtos.UpdateBookDto)
	err := ctx.BodyParser(&b)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "book body is invalid",
		})
	}

	id := ctx.Params("id")

	if id == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "id is required.",
		})
	}

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "invalid id.",
		})
	}

	collection := config.GetDBCollection("books")

	result, err := collection.UpdateOne(ctx.Context(), bson.M{"_id": objectId}, bson.M{"$set": b})

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "error while update book",
			"error":   err.Error(),
		})
	}

	return ctx.Status(200).JSON(fiber.Map{
		"success": true,
		"result":  result,
	})
}

func DeleteBook(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	if id == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "id is required.",
		})
	}
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "id is invalid.",
		})
	}

	collection := config.GetDBCollection("books")
	result, err := collection.DeleteOne(ctx.Context(), bson.M{"_id": objectId})

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "error while deleting book",
			"error":   err.Error(),
		})
	}

	return ctx.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
