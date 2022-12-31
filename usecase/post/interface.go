package post

import (
	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
)

//Reader interface
type Reader interface {
	Get(coupleID, userID, postID string) (entity.Post, error)
	List(id []entity.ID) ([]*entity.Post, error)
	GetByID(id entity.ID) (entity.Post, error)
	Comments(postID, userID string, skip int) ([]presentation.Comment, error)
	GetPosts(coupleID, userID entity.ID, postIDs []string) ([]presentation.Post, error)
	Explore(coupleIDs []entity.ID, userID entity.ID, country string, skip int) ([]presentation.Post, error)
	GetStats() interface{}
}

//Writer user writer
type Writer interface {
	Create(e *entity.Post) (entity.ID, error)
	Update(e *entity.Post) error
	Delete(coupleID, postID entity.ID) error
	AddComment(postID entity.ID, comment entity.Comment) error
	DeleteComment(postID, commentId string, userID entity.ID) error
	Like(postID, userID entity.ID) error
	UnLike(postID, userID entity.ID) error
	LikeComment(postID, commentID, userID entity.ID) error
	UnLikeComment(postID, commentID, userID entity.ID) error
	Edit(postID, coupleID entity.ID, edit entity.EditPost) error
	SetClosedComments(postID, coupleID entity.ID, state bool) error
}

//Repository interface
type Repository interface {
	Reader
	Writer
}

//Post use case
type UseCase interface {
	GetPost(coupleID, userID, postID string) (entity.Post, error)
	GetComments(postID, userID string, skip int) ([]presentation.Comment, error)
	GetPostByID(postID string) (entity.Post, error)
	CreatePost(e *entity.Post) (entity.ID, error)
	GetPosts(coupleID, userID entity.ID, postIDs []string) ([]presentation.Post, error)
	ListPosts(id []entity.ID) ([]*entity.Post, error)
	UpdatePost(e *entity.Post) error
	DeletePost(coupleID entity.ID, postID string) error
	NewComment(postID string, comment entity.Comment) error
	DeleteComment(postID, commentID string, userID entity.ID) error
	LikeComment(postID, commentID, userID entity.ID) error
	UnLikeComment(postID, commentID, userID entity.ID) error
	LikePost(postID, userID string) error
	UnLikePost(postID string, userID entity.ID) error
	EditPost(videoID string, coupleID entity.ID, edit entity.EditPost) error
	SetClosedComments(postID, coupleID entity.ID, state bool) error
	GetExplorePosts(coupleIDs []entity.ID, userID entity.ID, country string, skip int) ([]presentation.Post, error)
}
