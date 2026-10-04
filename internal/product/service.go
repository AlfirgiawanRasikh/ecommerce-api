package product

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetAll(
	ctx context.Context,
	query ProductQuery,
) (*ProductList, error) {
	return s.repository.GetAll(ctx, query)
}

func (s *Service) GetBySlug(
	ctx context.Context,
	slug string,
) (*Product, error) {
	return s.repository.GetBySlug(ctx, slug)
}

func (s *Service) Create(
	ctx context.Context,
	request CreateProductRequest,
) (*Product, error) {
	return s.repository.Create(ctx, request)
}

func (s *Service) Update(
	ctx context.Context,
	id int64,
	request UpdateProductRequest,
) (*Product, error) {
	return s.repository.Update(ctx, id, request)
}

func (s *Service) Delete(
	ctx context.Context,
	id int64,
) error {
	return s.repository.Delete(ctx, id)
}
