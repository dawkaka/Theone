package couple

import "github.com/dawkaka/theone/entity"

//Writer couple writer methods
type Writer interface {
	Create(couple entity.Couple) error
	Follower(userID, coupleID entity.ID) error
	Update(couple entity.Couple) error
}

//Reader couple reader methods
type Reader interface {
	Get(coupleName string) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) ([]entity.Post, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	Followers(coupleName string, skip int) ([]entity.Follower, error)
}

//Repository all couple methods
type Repository interface {
	Writer
	Reader
}

//Couple usecase
type UseCase interface {
	CreateCouple(userId, partnerrId, coupleName string) error
	UpdateCouple(couple entity.Couple) error
	GetCouple(coupleName string) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) ([]entity.Post, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	GetFollowers(coupleName string, skip int) ([]entity.Follower, error)
	NewFollower(userID, coupleID entity.ID) error
}
