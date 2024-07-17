package controllers

import (
	"context"
	"crud_mongo/config"
	dtos "crud_mongo/dtos"
	"crud_mongo/models"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetBooks(ctx *fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	// Try to get from cache
	cacheKey := fmt.Sprintf("books:%d:%d", page, limit)

	// config.RedisDB.Set(context.Background(), cacheKey, string(bookBytes), 10*time.Minute)


	cachedBooks, err := config.RedisDB.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var books []models.Book
		if err := json.Unmarshal([]byte(cachedBooks), &books); err == nil {
			return ctx.JSON(fiber.Map{
				"success": true,
				"data":    books,
			})
		}
	}

	collection := config.GetDBCollection("books")

	// Create a context with a 10-second timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	books := make([]models.Book, 0)
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	res, err := collection.Find(ctxTimeout, bson.M{}, findOptions)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	defer res.Close(ctxTimeout)

	for res.Next(ctxTimeout) {
		var book models.Book
		err := res.Decode(&book)
		if err != nil {
			return ctx.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}
		books = append(books, book)
	}

	if err := res.Err(); err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Cache the result
	bookBytes, err := json.Marshal(books)
	if err == nil {
		config.RedisDB.Set(context.Background(), cacheKey, string(bookBytes), 10*time.Minute)
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"Number of books": len(books),
		"data":    books,
	})
}

func GetBook(ctx *fiber.Ctx) error {
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
			"message": "id is invalid",
		})
	}

	// Try to get the book from the cache
	cacheKey := "book:" + id
	cachedBook, err := config.RedisDB.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var book models.Book
		if err := json.Unmarshal([]byte(cachedBook), &book); err == nil {
			return ctx.Status(200).JSON(fiber.Map{
				"success": true,
				"data":    book,
			})
		}
	}

	collection := config.GetDBCollection("books")

	// Create a context with a 10-second timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	book := models.Book{}
	err = collection.FindOne(ctxTimeout, bson.M{"_id": objectId}).Decode(&book)
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "book not found",
		})
	}

	// Cache the book
	bookBytes, err := json.Marshal(book)
	if err == nil {
		config.RedisDB.Set(context.Background(), cacheKey, string(bookBytes), 10*time.Minute)
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

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := config.GetDBCollection("books")
	result, err := collection.InsertOne(ctxTimeout, b)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "failed to create book",
			"error":   err.Error(),
		})
	}

	config.RedisDB.Del(ctx.Context(), "books")

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

	// Create a context with a 10-second timeout
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctxTimeout, bson.M{"_id": objectId}, bson.M{"$set": b})

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "error while update book",
			"error":   err.Error(),
		})
	}

	// cache
	cacheKey := "book:" + id
	config.RedisDB.Del(context.Background(), cacheKey)

	iter := config.RedisDB.Scan(context.Background(), 0, "books:*", 0).Iterator()
	for iter.Next(context.Background()) {
		config.RedisDB.Del(context.Background(), iter.Val())
	}
	if err := iter.Err(); err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "error while invalidating cache",
			"error":   err.Error(),
		})
	}

	// Optionally, get the updated book and cache it
	var updatedBook models.Book
	if err := collection.FindOne(ctxTimeout, bson.M{"_id": objectId}).Decode(&updatedBook); err == nil {
		bookBytes, err := json.Marshal(updatedBook)
		if err == nil {
			config.RedisDB.Set(context.Background(), cacheKey, string(bookBytes), 10*time.Minute)
		}
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

	// Delete the book from MongoDB
	result, err := collection.DeleteOne(ctx.Context(), bson.M{"_id": objectId})
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "error while deleting book",
			"error":   err.Error(),
		})
	}

	// Invalidate the cache for the deleted book
	cacheKey := "book:" + id
	config.RedisDB.Del(ctx.Context(), cacheKey)

	return ctx.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
