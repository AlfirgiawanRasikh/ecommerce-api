package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"ecommerce-api/internal/database"
)

type ProductImage struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	IsPrimary bool   `json:"isPrimary"`
	SortOrder *int   `json:"sortOrder,omitempty"`
}

type ProductColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

type ProductVariant struct {
	ID    string       `json:"id"`
	SKU   string       `json:"sku"`
	Size  string       `json:"size"`
	Color ProductColor `json:"color"`
	Price float64      `json:"price"`
	Stock int          `json:"stock"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Product struct {
	ID          string           `json:"id"`
	Slug        string           `json:"slug"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	BasePrice   float64          `json:"basePrice"`
	WeightGrams int              `json:"weightGrams"`
	Category    Category         `json:"category"`
	Brand       Brand            `json:"brand"`
	Images      []ProductImage   `json:"images"`
	Variants    []ProductVariant `json:"variants"`
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env:", err)
	}

	// Connect to PostgreSQL
	db, err := database.Connect()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	// Read products.json
	jsonPath := `D:\projectlatihan\frontend-ecommerce\src\features\products\data\products.json`

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatal("failed to read products.json:", err)
	}

	var products []Product

	// Remove UTF-8 BOM if present.
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}

	if err := json.Unmarshal(data, &products); err != nil {
		log.Fatal("failed to parse products.json:", err)
	}

	fmt.Printf("Found %d products\n", len(products))

	ctx := context.Background()

	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal("failed to begin transaction:", err)
	}

	defer tx.Rollback(ctx)

	categoryIDs := make(map[string]int64)
	brandIDs := make(map[string]int64)

	for _, product := range products {

		// ============================================
		// CATEGORY
		// ============================================

		categoryID, exists := categoryIDs[product.Category.Slug]

		if !exists {
			err := tx.QueryRow(
				ctx,
				`
				INSERT INTO categories (name, slug)
				VALUES ($1, $2)
				ON CONFLICT (slug)
				DO UPDATE SET name = EXCLUDED.name
				RETURNING id
				`,
				product.Category.Name,
				product.Category.Slug,
			).Scan(&categoryID)

			if err != nil {
				log.Fatal("failed to insert category:", err)
			}

			categoryIDs[product.Category.Slug] = categoryID
		}

		// ============================================
		// BRAND
		// ============================================

		brandID, exists := brandIDs[product.Brand.Slug]

		if !exists {
			err := tx.QueryRow(
				ctx,
				`
				INSERT INTO brands (name, slug)
				VALUES ($1, $2)
				ON CONFLICT (slug)
				DO UPDATE SET name = EXCLUDED.name
				RETURNING id
				`,
				product.Brand.Name,
				product.Brand.Slug,
			).Scan(&brandID)

			if err != nil {
				log.Fatal("failed to insert brand:", err)
			}

			brandIDs[product.Brand.Slug] = brandID
		}

		// ============================================
		// PRODUCT
		// ============================================

		var productID int64

		err = tx.QueryRow(
			ctx,
			`
			INSERT INTO products (
				name,
				slug,
				description,
				base_price,
				weight_grams,
				category_id,
				brand_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (slug)
			DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				base_price = EXCLUDED.base_price,
				weight_grams = EXCLUDED.weight_grams,
				category_id = EXCLUDED.category_id,
				brand_id = EXCLUDED.brand_id,
				updated_at = NOW()
			RETURNING id
			`,
			product.Name,
			product.Slug,
			product.Description,
			product.BasePrice,
			product.WeightGrams,
			categoryID,
			brandID,
		).Scan(&productID)

		if err != nil {
			log.Fatal("failed to insert product:", err)
		}

		// ============================================
		// PRODUCT IMAGES
		// ============================================

		for _, image := range product.Images {
			sortOrder := 0

			if image.SortOrder != nil {
				sortOrder = *image.SortOrder
			}

			_, err = tx.Exec(
				ctx,
				`
				INSERT INTO product_images (
					product_id,
					url,
					alt,
					is_primary,
					sort_order
				)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT DO NOTHING
				`,
				productID,
				image.URL,
				image.Alt,
				image.IsPrimary,
				sortOrder,
			)

			if err != nil {
				log.Fatal("failed to insert product image:", err)
			}
		}

		// ============================================
		// PRODUCT VARIANTS
		// ============================================

		for _, variant := range product.Variants {
			_, err = tx.Exec(
				ctx,
				`
				INSERT INTO product_variants (
					product_id,
					sku,
					size,
					color_name,
					color_hex,
					price,
					stock
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (sku)
				DO UPDATE SET
					product_id = EXCLUDED.product_id,
					size = EXCLUDED.size,
					color_name = EXCLUDED.color_name,
					color_hex = EXCLUDED.color_hex,
					price = EXCLUDED.price,
					stock = EXCLUDED.stock,
					updated_at = NOW()
				`,
				productID,
				variant.SKU,
				variant.Size,
				variant.Color.Name,
				variant.Color.Hex,
				variant.Price,
				variant.Stock,
			)

			if err != nil {
				log.Fatal("failed to insert product variant:", err)
			}
		}
	}

	// ============================================
	// COMMIT
	// ============================================

	if err := tx.Commit(ctx); err != nil {
		log.Fatal("failed to commit transaction:", err)
	}

	fmt.Println("Seed completed successfully!")
}
