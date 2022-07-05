package post

import "github.com/dawkaka/theone/entity"

//Post service
type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) CreatePost(e *entity.Post) (entity.ID, error) {
	id, err := s.repo.Create(e)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (s *Service) ListPosts(userName string) ([]*entity.Post, error) {
	posts, err := s.repo.List(userName)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *Service) UpdatePost(e *entity.Post) error {
	return s.repo.Update(e)
}

func (s *Service) DeletePost(id entity.ID) error {
	return s.repo.Delete(id)
}
