package post

import "github.com/dawkaka/theone/entity"

//Reader interface
type Reader interface {
	Get(coupleName, postID string) (*entity.Post, error)
	List(id []entity.ID) ([]*entity.Post, error)
}

//Writer user writer
type Writer interface {
	Create(e *entity.Post) (entity.ID, error)
	Update(e *entity.Post) error
	Delete(id entity.ID) error
}

//Repository interface
type Repository interface {
	Reader
	Writer
}

//Post use case
type UseCase interface {
	GetPost(coupleName, postID string) (*entity.Post, error)
	CreatePost(e *entity.Post) (entity.ID, error)
	ListPosts(id []entity.ID) ([]*entity.Post, error)
	UpdatePost(e *entity.Post) error
	DeletePost(id entity.ID) error
}
