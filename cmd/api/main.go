package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ecommerce-api/internal/category"
	"ecommerce-api/internal/database"
	"ecommerce-api/internal/product"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env:", err)
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	defer db.Close()

	productRepository := product.NewRepository(db)
	productService := product.NewService(productRepository)
	productHandler := product.NewHandler(productService)

	categoryRepository := category.NewRepository(db)
	categoryService := category.NewService(categoryRepository)
	categoryHandler := category.NewHandler(categoryService)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5000",
			"http://127.0.0.1:5000",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.GET("/health/db", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	router.GET("/products", productHandler.GetAll)
	router.GET("/products/:slug", productHandler.GetBySlug)

	router.POST("/products", productHandler.Create)
	router.PUT("/products/:id", productHandler.Update)
	router.DELETE("/products/:id", productHandler.Delete)

	router.GET("/categories", categoryHandler.GetAll)
	router.GET("/categories/:id", categoryHandler.GetByID)
	router.POST("/categories", categoryHandler.Create)
	router.PUT("/categories/:id", categoryHandler.Update)
	router.DELETE("/categories/:id", categoryHandler.Delete)

	log.Println("Server running on http://localhost:8080")

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
