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

func (s *Service) GetComments(videoId string, skip int) ([]entity.Comment, error) {
	return s.repo.Comments(videoId, skip)
}
func (s *Service) GetVideoByID(id string) (entity.Video, error) {
	return s.repo.GetByID(id)
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

func (s *Service) DeleteVideo(coupleID entity.ID, videoID string) error {
	newVideoID, err := entity.StringToID(videoID)
	if err != nil {
		return entity.ErrInvalidID
	}
	return s.repo.Delete(coupleID, newVideoID)
}

func (s *Service) NewComment(videoID string, comment entity.Comment) error {
	id, err := entity.StringToID(videoID)
	if err != nil {
		return err
	}
	return s.repo.AddComment(id, comment)
}

func (s *Service) DeleteComment(videoID, commentID string, userID entity.ID) error {
	return s.repo.DeleteComment(videoID, commentID, userID)
}

func (s *Service) UnLikeVideo(videoID string, userID entity.ID) error {
	ID, err := entity.StringToID(videoID)
	if err != nil {
		return err
	}
	return s.repo.UnLike(ID, userID)
}

func (s *Service) EditCaption(videoID string, coupleID entity.ID, newCaption string) error {
	ID, err := entity.StringToID(videoID)
	if err != nil {
		return err
	}
	return s.repo.Edit(ID, coupleID, newCaption)
}

func (s *Service) LikeVideo(videoID, userID string) error {
	id, err := entity.StringToID(userID)
	if err != nil {
		return err
	}
	pID, err := entity.StringToID(videoID)
	if err != nil {
		return err
	}
	return s.repo.Like(pID, id)
}
