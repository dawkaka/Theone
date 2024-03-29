package post

import (
	"errors"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
)

//Post service
type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetPost(coupleID, userID, postID string) (entity.Post, error) {
	return s.repo.Get(coupleID, userID, postID)
}

func (s *Service) GetPostByID(id string) (entity.Post, error) {
	ID, err := entity.StringToID(id)
	if err != nil {
		return entity.Post{}, errors.New("parsing ID: not a mongodb id")
	}
	return s.repo.GetByID(ID)
}

func (s *Service) GetComments(postID, userID string, skip int) ([]presentation.Comment, error) {
	return s.repo.Comments(postID, userID, skip)
}

func (s *Service) CreatePost(p *entity.Post) (entity.ID, error) {
	id, err := s.repo.Create(p)
	return id, err
}

func (s *Service) NewComment(postID string, comment entity.Comment) error {
	id, err := entity.StringToID(postID)
	if err != nil {
		return err
	}
	return s.repo.AddComment(id, comment)
}

func (s *Service) DeleteComment(postID, commentID string, userID entity.ID) error {
	return s.repo.DeleteComment(postID, commentID, userID)
}

func (s *Service) LikeComment(postID, commentID, userID entity.ID) error {
	return s.repo.LikeComment(postID, commentID, userID)
}
func (s *Service) UnLikeComment(postID, commentID, userID entity.ID) error {
	return s.repo.UnLikeComment(postID, commentID, userID)
}

func (s *Service) LikePost(postID, userID string) error {
	id, err := entity.StringToID(userID)
	if err != nil {
		return err
	}
	pID, err := entity.StringToID(postID)
	if err != nil {
		return err
	}
	return s.repo.Like(pID, id)
}

func (s *Service) UnLikePost(postID string, userID entity.ID) error {
	ID, err := entity.StringToID(postID)
	if err != nil {
		return err
	}
	return s.repo.UnLike(ID, userID)
}

func (s *Service) EditPost(postID string, coupleID entity.ID, edit entity.EditPost) error {
	ID, err := entity.StringToID(postID)
	if err != nil {
		return err
	}
	return s.repo.Edit(ID, coupleID, edit)
}

func (s *Service) ListPosts(id []entity.ID) ([]*entity.Post, error) {
	posts, err := s.repo.List(id)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *Service) UpdatePost(e *entity.Post) error {
	return s.repo.Update(e)
}

func (s *Service) DeletePost(coupleID entity.ID, postID string) error {
	newPostID, err := entity.StringToID(postID)
	if err != nil {
		return err
	}
	return s.repo.Delete(coupleID, newPostID)
}

func (s *Service) GetPosts(coupleID, userID entity.ID, postIDs []string) ([]presentation.Post, error) {
	return s.repo.GetPosts(coupleID, userID, postIDs)
}

func (s *Service) SetClosedComments(postID, coupleID entity.ID, state bool) error {
	return s.repo.SetClosedComments(postID, coupleID, state)
}

func (s *Service) GetExplorePosts(coupleIDs []entity.ID, userID entity.ID, country string, skip int) ([]presentation.Post, error) {
	return s.repo.Explore(coupleIDs, userID, country, skip)
}
