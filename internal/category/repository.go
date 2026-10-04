package category

import (
	"context"

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

func (r *Repository) GetAll(ctx context.Context) ([]Category, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, slug
		FROM categories
		ORDER BY name ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)

	for rows.Next() {
		var category Category

		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
		); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (*Category, error) {
	var category Category

	err := r.db.QueryRow(ctx, `
		SELECT id, name, slug
		FROM categories
		WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
	)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *Repository) Create(
	ctx context.Context,
	request CreateCategoryRequest,
) (*Category, error) {
	var id int64

	err := r.db.QueryRow(ctx, `
		INSERT INTO categories (
			name,
			slug
		)
		VALUES ($1, $2)
		RETURNING id
	`,
		request.Name,
		request.Slug,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Update(
	ctx context.Context,
	id int64,
	request UpdateCategoryRequest,
) (*Category, error) {
	var category Category

	err := r.db.QueryRow(ctx, `
		UPDATE categories
		SET
			name = $1,
			slug = $2,
			updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, slug
	`,
		request.Name,
		request.Slug,
		id,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
	)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id int64,
) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM categories
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
