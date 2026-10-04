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
