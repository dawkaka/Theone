package couple

import (
	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
)

//Writer couple writer methods
type Writer interface {
	Create(couple entity.Couple) (entity.ID, error)
	Follower(userID, coupleID entity.ID) error
	Unfollow(coupleID, userID entity.ID) error
	Update(coupleID entity.ID, update entity.UpdateCouple) error
	BreakUp(coupleID entity.ID) error
	UpdateProfilePic(fileName string, coupleID entity.ID) error
	UpdateCoverPic(fileName string, coupleID entity.ID) error
	ChangeName(coupleID entity.ID, coupleName string) error
}

//Reader couple reader methods
type Reader interface {
	Get(coupleName string) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) ([]entity.Post, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	Followers(coupleName string, skip int) ([]entity.ID, error)
	List(coupleIDs []entity.ID) ([]presentation.CouplePreview, error)
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
	GetFollowers(coupleName string, skip int) ([]entity.ID, error)
	NewFollower(userID, coupleID entity.ID) error
	RemoveFollower(coupleID, userID entity.ID) error
	UpdateCoupleProfilePic(fileName string, coupleID entity.ID) error
	UpdateCoupleCoverPic(fileName string, coupleID entity.ID) error
	ChangeCoupleName(coupleID entity.ID, coupleName string) error
	BreakUp(coupleID entity.ID) error
	ListCouple(coupleIDs []entity.ID) ([]presentation.CouplePreview, error)
}
