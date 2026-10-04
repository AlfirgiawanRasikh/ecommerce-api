package product

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

type ProductQuery struct {
	Search   string
	Category string
	Brand    string
	Size     string
	Color    string
	MinPrice *float64
	MaxPrice *float64
	Sort     string
	Page     int
	Limit    int
}

type ProductList struct {
	Products   []Product
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

func (r *Repository) GetAll(
	ctx context.Context,
	query ProductQuery,
) (*ProductList, error) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	argIndex := 1

	if query.Search != "" {
		conditions = append(
			conditions,
			fmt.Sprintf(`
				(
					p.name ILIKE $%d
					OR p.description ILIKE $%d
					OR c.name ILIKE $%d
					OR b.name ILIKE $%d
					OR EXISTS (
						SELECT 1
						FROM product_variants pv
						WHERE pv.product_id = p.id
						AND pv.sku ILIKE $%d
					)
				)
			`, argIndex, argIndex, argIndex, argIndex, argIndex),
		)

		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	if query.Category != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("c.slug = $%d", argIndex),
		)

		args = append(args, query.Category)
		argIndex++
	}

	if query.Brand != "" {
		conditions = append(
			conditions,
			fmt.Sprintf("b.slug = $%d", argIndex),
		)

		args = append(args, query.Brand)
		argIndex++
	}

	if query.Size != "" {
		conditions = append(
			conditions,
			fmt.Sprintf(`
				EXISTS (
					SELECT 1
					FROM product_variants pv
					WHERE pv.product_id = p.id
					AND pv.size = $%d
				)
			`, argIndex),
		)

		args = append(args, query.Size)
		argIndex++
	}

	if query.Color != "" {
		conditions = append(
			conditions,
			fmt.Sprintf(`
				EXISTS (
					SELECT 1
					FROM product_variants pv
					WHERE pv.product_id = p.id
					AND LOWER(pv.color_name) = LOWER($%d)
				)
			`, argIndex),
		)

		args = append(args, query.Color)
		argIndex++
	}

	if query.MinPrice != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("p.base_price >= $%d", argIndex),
		)

		args = append(args, *query.MinPrice)
		argIndex++
	}

	if query.MaxPrice != nil {
		conditions = append(
			conditions,
			fmt.Sprintf("p.base_price <= $%d", argIndex),
		)

		args = append(args, *query.MaxPrice)
		argIndex++
	}

	whereClause := ""

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	orderClause := "p.id DESC"

	switch query.Sort {
	case "price-asc":
		orderClause = "p.base_price ASC, p.id DESC"

	case "price-desc":
		orderClause = "p.base_price DESC, p.id DESC"

	case "name-asc":
		orderClause = "p.name ASC, p.id DESC"

	case "name-desc":
		orderClause = "p.name DESC, p.id DESC"

	case "newest":
		orderClause = "p.created_at DESC, p.id DESC"
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM products p

		JOIN categories c
			ON c.id = p.category_id

		JOIN brands b
			ON b.id = p.brand_id

		%s
	`, whereClause)

	var total int

	err := r.db.QueryRow(
		ctx,
		countQuery,
		args...,
	).Scan(&total)

	if err != nil {
		return nil, err
	}

	offset := (query.Page - 1) * query.Limit

	dataQuery := fmt.Sprintf(`
		SELECT
			p.id,
			p.name,
			p.slug,
			p.description,
			p.base_price,
			p.weight_grams,

			c.id,
			c.name,
			c.slug,

			b.id,
			b.name,
			b.slug

		FROM products p

		JOIN categories c
			ON c.id = p.category_id

		JOIN brands b
			ON b.id = p.brand_id

		%s

		ORDER BY %s

		LIMIT $%d
		OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)

	dataArgs := append(
		append([]any{}, args...),
		query.Limit,
		offset,
	)

	rows, err := r.db.Query(
		ctx,
		dataQuery,
		dataArgs...,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	products := make([]Product, 0, query.Limit)

	for rows.Next() {
		var product Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Slug,
			&product.Description,
			&product.BasePrice,
			&product.WeightGrams,

			&product.Category.ID,
			&product.Category.Name,
			&product.Category.Slug,

			&product.Brand.ID,
			&product.Brand.Name,
			&product.Brand.Slug,
		)

		if err != nil {
			return nil, err
		}

		images, err := r.getImages(
			ctx,
			product.ID,
		)

		if err != nil {
			return nil, err
		}

		variants, err := r.getVariants(
			ctx,
			product.ID,
		)

		if err != nil {
			return nil, err
		}

		product.Images = images
		product.Variants = variants

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := 0

	if total > 0 {
		totalPages = (total + query.Limit - 1) / query.Limit
	}

	return &ProductList{
		Products:   products,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *Repository) GetBySlug(
	ctx context.Context,
	slug string,
) (*Product, error) {
	var product Product

	err := r.db.QueryRow(ctx, `
		SELECT
			p.id,
			p.name,
			p.slug,
			p.description,
			p.base_price,
			p.weight_grams,

			c.id,
			c.name,
			c.slug,

			b.id,
			b.name,
			b.slug

		FROM products p

		JOIN categories c
			ON c.id = p.category_id

		JOIN brands b
			ON b.id = p.brand_id

		WHERE p.slug = $1
	`, slug).Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.BasePrice,
		&product.WeightGrams,

		&product.Category.ID,
		&product.Category.Name,
		&product.Category.Slug,

		&product.Brand.ID,
		&product.Brand.Name,
		&product.Brand.Slug,
	)

	if err != nil {
		return nil, err
	}

	images, err := r.getImages(ctx, product.ID)

	if err != nil {
		return nil, err
	}

	variants, err := r.getVariants(ctx, product.ID)

	if err != nil {
		return nil, err
	}

	product.Images = images
	product.Variants = variants

	return &product, nil
}

func (r *Repository) getImages(
	ctx context.Context,
	productID int64,
) ([]ProductImage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			url,
			alt,
			is_primary,
			sort_order
		FROM product_images
		WHERE product_id = $1
		ORDER BY sort_order, id
	`, productID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	images := make([]ProductImage, 0)

	for rows.Next() {
		var image ProductImage

		err := rows.Scan(
			&image.ID,
			&image.URL,
			&image.Alt,
			&image.IsPrimary,
			&image.SortOrder,
		)

		if err != nil {
			return nil, err
		}

		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func (r *Repository) getVariants(
	ctx context.Context,
	productID int64,
) ([]ProductVariant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			sku,
			size,
			color_name,
			color_hex,
			price,
			stock
		FROM product_variants
		WHERE product_id = $1
		ORDER BY id
	`, productID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	variants := make([]ProductVariant, 0)

	for rows.Next() {
		var variant ProductVariant

		err := rows.Scan(
			&variant.ID,
			&variant.SKU,
			&variant.Size,
			&variant.Color.Name,
			&variant.Color.Hex,
			&variant.Price,
			&variant.Stock,
		)

		if err != nil {
			return nil, err
		}

		variants = append(variants, variant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return variants, nil
}

func ParseFloat(value string) (*float64, error) {
	if value == "" {
		return nil, nil
	}

	number, err := strconv.ParseFloat(value, 64)

	if err != nil {
		return nil, err
	}

	return &number, nil
}

func (r *Repository) Create(
	ctx context.Context,
	request CreateProductRequest,
) (*Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var productID int64

	err = tx.QueryRow(ctx, `
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
		RETURNING id
	`,
		request.Name,
		request.Slug,
		request.Description,
		request.BasePrice,
		request.WeightGrams,
		request.CategoryID,
		request.BrandID,
	).Scan(&productID)

	if err != nil {
		return nil, err
	}

	for _, image := range request.Images {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_images (
				product_id,
				url,
				alt,
				is_primary,
				sort_order
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
			productID,
			image.URL,
			image.Alt,
			image.IsPrimary,
			image.SortOrder,
		)

		if err != nil {
			return nil, err
		}
	}

	for _, variant := range request.Variants {
		_, err = tx.Exec(ctx, `
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
		`,
			productID,
			variant.SKU,
			variant.Size,
			variant.ColorName,
			variant.ColorHex,
			variant.Price,
			variant.Stock,
		)

		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, productID)
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (*Product, error) {
	var product Product

	err := r.db.QueryRow(ctx, `
		SELECT
			p.id,
			p.name,
			p.slug,
			p.description,
			p.base_price,
			p.weight_grams,
			c.id,
			c.name,
			c.slug,
			b.id,
			b.name,
			b.slug
		FROM products p
		JOIN categories c ON c.id = p.category_id
		JOIN brands b ON b.id = p.brand_id
		WHERE p.id = $1
	`, id).Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.BasePrice,
		&product.WeightGrams,
		&product.Category.ID,
		&product.Category.Name,
		&product.Category.Slug,
		&product.Brand.ID,
		&product.Brand.Name,
		&product.Brand.Slug,
	)

	if err != nil {
		return nil, err
	}

	images, err := r.getImages(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	variants, err := r.getVariants(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	product.Images = images
	product.Variants = variants

	return &product, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id int64,
	request UpdateProductRequest,
) (*Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var productID int64

	err = tx.QueryRow(ctx, `
		UPDATE products
		SET
			name = $1,
			slug = $2,
			description = $3,
			base_price = $4,
			weight_grams = $5,
			category_id = $6,
			brand_id = $7,
			updated_at = NOW()
		WHERE id = $8
		RETURNING id
	`,
		request.Name,
		request.Slug,
		request.Description,
		request.BasePrice,
		request.WeightGrams,
		request.CategoryID,
		request.BrandID,
		id,
	).Scan(&productID)

	if err != nil {
		return nil, err
	}

	// Hapus images lama
	_, err = tx.Exec(ctx, `
		DELETE FROM product_images
		WHERE product_id = $1
	`, productID)

	if err != nil {
		return nil, err
	}

	// Insert images baru
	for _, image := range request.Images {
		_, err = tx.Exec(ctx, `
			INSERT INTO product_images (
				product_id,
				url,
				alt,
				is_primary,
				sort_order
			)
			VALUES ($1, $2, $3, $4, $5)
		`,
			productID,
			image.URL,
			image.Alt,
			image.IsPrimary,
			image.SortOrder,
		)

		if err != nil {
			return nil, err
		}
	}

	// Hapus variants lama
	_, err = tx.Exec(ctx, `
		DELETE FROM product_variants
		WHERE product_id = $1
	`, productID)

	if err != nil {
		return nil, err
	}

	// Insert variants baru
	for _, variant := range request.Variants {
		_, err = tx.Exec(ctx, `
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
		`,
			productID,
			variant.SKU,
			variant.Size,
			variant.ColorName,
			variant.ColorHex,
			variant.Price,
			variant.Stock,
		)

		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, productID)
}

func (r *Repository) Delete(
	ctx context.Context,
	id int64,
) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM products
		WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
