package category

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]Category, error) {
	return s.repository.GetAll(ctx)
}

func (s *Service) GetByID(
	ctx context.Context,
	id int64,
) (*Category, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Create(
	ctx context.Context,
	request CreateCategoryRequest,
) (*Category, error) {
	return s.repository.Create(ctx, request)
}

func (s *Service) Update(
	ctx context.Context,
	id int64,
	request UpdateCategoryRequest,
) (*Category, error) {
	return s.repository.Update(ctx, id, request)
}

func (s *Service) Delete(
	ctx context.Context,
	id int64,
) error {
	return s.repository.Delete(ctx, id)
}
