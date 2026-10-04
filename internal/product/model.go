package product

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Brand struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProductImage struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	IsPrimary bool   `json:"isPrimary"`
	SortOrder int    `json:"sortOrder"`
}

type ProductColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
}

type ProductVariant struct {
	ID    int64        `json:"id"`
	SKU   string       `json:"sku"`
	Size  string       `json:"size"`
	Color ProductColor `json:"color"`
	Price float64      `json:"price"`
	Stock int          `json:"stock"`
}

type Product struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description"`
	BasePrice   float64          `json:"basePrice"`
	WeightGrams int              `json:"weightGrams"`
	Category    Category         `json:"category"`
	Brand       Brand            `json:"brand"`
	Images      []ProductImage   `json:"images"`
	Variants    []ProductVariant `json:"variants"`
}

type CreateProductRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Slug        string                 `json:"slug" binding:"required"`
	Description string                 `json:"description"`
	BasePrice   float64                `json:"basePrice" binding:"gte=0"`
	WeightGrams int                    `json:"weightGrams" binding:"gte=0"`
	CategoryID  int64                  `json:"categoryId" binding:"required"`
	BrandID     int64                  `json:"brandId" binding:"required"`
	Images      []CreateProductImage   `json:"images"`
	Variants    []CreateProductVariant `json:"variants"`
}

type CreateProductImage struct {
	URL       string `json:"url" binding:"required"`
	Alt       string `json:"alt"`
	IsPrimary bool   `json:"isPrimary"`
	SortOrder int    `json:"sortOrder"`
}

type CreateProductVariant struct {
	SKU       string  `json:"sku" binding:"required"`
	Size      string  `json:"size" binding:"required"`
	ColorName string  `json:"colorName" binding:"required"`
	ColorHex  string  `json:"colorHex" binding:"required"`
	Price     float64 `json:"price" binding:"gte=0"`
	Stock     int     `json:"stock" binding:"gte=0"`
}

type UpdateProductRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Slug        string                 `json:"slug" binding:"required"`
	Description string                 `json:"description"`
	BasePrice   float64                `json:"basePrice" binding:"gte=0"`
	WeightGrams int                    `json:"weightGrams" binding:"gte=0"`
	CategoryID  int64                  `json:"categoryId" binding:"required"`
	BrandID     int64                  `json:"brandId" binding:"required"`
	Images      []CreateProductImage   `json:"images"`
	Variants    []CreateProductVariant `json:"variants"`
}
