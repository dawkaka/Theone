package post

import "github.com/dawkaka/theone/entity"

//Reader interface
type Reader interface {
	Get(coupleID, postID string) (*entity.Post, error)
	List(id []entity.ID) ([]*entity.Post, error)
	GetByID(id entity.ID) (entity.Post, error)
	Comments(postID string, skip int) ([]entity.Comment, error)
}

//Writer user writer
type Writer interface {
	Create(e *entity.Post) (entity.ID, error)
	Update(e *entity.Post) error
	Delete(id entity.ID) error
	AddComment(postID entity.ID, comment entity.Comment) error
	DeleteComment(postID, commentId string, userID entity.ID) error
	Like(postID, userID entity.ID) error
	UnLike(postID, userID entity.ID) error
	Edit(postID, coupleID entity.ID, newCaption string) error
}

//Repository interface
type Repository interface {
	Reader
	Writer
}

//Post use case
type UseCase interface {
	GetPost(coupleID, postID string) (*entity.Post, error)
	GetComments(postID string, skip int) ([]entity.Comment, error)
	GetPostByID(postID string) (entity.Post, error)
	CreatePost(e *entity.Post) (entity.ID, error)
	ListPosts(id []entity.ID) ([]*entity.Post, error)
	UpdatePost(e *entity.Post) error
	DeletePost(id entity.ID) error
	NewComment(postID string, comment entity.Comment) error
	DeleteComment(postID, commentID string, userID entity.ID) error
	LikePost(postID, userID string) error
	UnLikePost(postID string, userID entity.ID) error
	EditCaption(videoID string, coupleID entity.ID, newCaption string) error
}
