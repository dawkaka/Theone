package couple

import "github.com/dawkaka/theone/entity"

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetCouple(coupleName string) (entity.Couple, error) {
	return s.repo.Get(coupleName)
}

func (s *Service) GetCouplePosts(coupleName string, skip int) ([]entity.Post, error) {
	return s.repo.GetCouplePosts(coupleName, skip)
}

func (s *Service) GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error) {
	return s.repo.GetCoupleVideos(coupleName, skip)
}

func (s *Service) CreateCouple(couple entity.Couple) error {
	return s.repo.Create(couple)
}

func (s *Service) UpdateCouple(couple entity.Couple) error {

	return s.repo.Update(couple)
}
