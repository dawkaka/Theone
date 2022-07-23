package couple

import "github.com/dawkaka/theone/entity"

//Writer couple writer methods
type Writer interface {
	Create(couple entity.Couple) (entity.ID, error)
	Follower(userID, coupleID entity.ID) error
	Unfollow(userID, coupleID entity.ID) error
	Update(coupleID entity.ID, update entity.UpdateCouple) error
}

//Reader couple reader methods
type Reader interface {
	Get(coupleName string) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) ([]entity.Post, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	Followers(coupleName string, skip int) ([]entity.Follower, error)
	UpdateProfilePic(fileName string, coupleID entity.ID) error
	UpdateCoverPic(fileName string, coupleID entity.ID) error
	ChangeName(coupleID entity.ID, coupleName string) error
}

//Repository all couple methods
type Repository interface {
	Writer
	Reader
}

//Couple usecase
type UseCase interface {
	CreateCouple(userId, partnerrId, coupleName string) (entity.ID, error)
	UpdateCouple(coupleID entity.ID, update entity.UpdateCouple) error
	GetCouple(coupleName string) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) ([]entity.Post, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	GetFollowers(coupleName string, skip int) ([]entity.Follower, error)
	NewFollower(userID, coupleID entity.ID) error
	RemoveFollower(userID, coupleID entity.ID) error
	UpdateCoupleProfilePic(fileName string, coupleID entity.ID) error
	UpdateCoupleCoverPic(fileName string, coupleID entity.ID) error
	ChangeCoupleName(coupleID entity.ID, coupleName string) error
}
