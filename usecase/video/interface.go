package video

import "github.com/dawkaka/theone/entity"

//Reader interface
type Reader interface {
	Get(coupleID, videoID string) (*entity.Video, error)
	List(videos []entity.ID) ([]*entity.Video, error)
}

//Writer user writer
type Writer interface {
	Create(e *entity.Video) (entity.ID, error)
	Update(e *entity.Video) error
	Delete(id entity.ID) error
	AddComment(videoID entity.ID, comment entity.Comment) error
	Like(videoID, userID entity.ID) error
}

//Repository interface
type Repository interface {
	Reader
	Writer
}

//UseCase interface
type UseCase interface {
	GetVideo(coupleID, videoID string) (*entity.Video, error)
	ListVideos(ids []entity.ID) ([]*entity.Video, error)
	CreateVideo(video *entity.Video) (entity.ID, error)
	UpdateVideo(e *entity.Video) error
	DeleteVideo(id entity.ID) error
	NewComment(videoID string, comment entity.Comment) error
	LikeVideo(videoID, userID string) error
}
