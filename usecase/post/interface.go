package post

import "github.com/dawkaka/theone/entity"

//Reader interface
type Reader interface {
	Get(id entity.ID) (*entity.Post, error)
	List(userName string) ([]*entity.Post, error)
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
	CreatePost(e *entity.Post) (entity.ID, error)
	ListPosts() []*entity.Post
	UpdatePost(e *entity.Post) error
	DeletePost(id entity.ID) error
}
