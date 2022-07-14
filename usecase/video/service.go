package video

import "github.com/dawkaka/theone/entity"

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetVideo(coupleID, videoID string) (*entity.Video, error) {
	return s.repo.Get(coupleID, videoID)
}

func (s *Service) ListVideos(ids []entity.ID) ([]*entity.Video, error) {
	posts, err := s.repo.List(ids)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *Service) CreateVideo(e *entity.Video) (entity.ID, error) {
	id, err := s.repo.Create(e)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (s *Service) UpdateVideo(e *entity.Video) error {
	return s.repo.Update(e)
}

func (s *Service) DeleteVideo(id entity.ID) error {
	return s.repo.Delete(id)
}
