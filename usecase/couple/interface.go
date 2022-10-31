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
	MakeUp(coupleID entity.ID) error
	Dated(userID, partnerID entity.ID) (entity.ID, error)
	AddPost(coupleID entity.ID, postID string) error
	RemovePost(coupleID entity.ID, postID string) error
	UpdateStatus(coupleID entity.ID, married bool) error
	Block(coupleID, userID entity.ID) error
}

//Reader couple reader methods
type Reader interface {
	Get(coupleName string, userID entity.ID) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) (entity.Couple, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	Search(query string, userID entity.ID) ([]presentation.CouplePreview, error)
	Followers(coupleName string, skip int) ([]entity.ID, error)
	List(coupleIDs []entity.ID, userID entity.ID) ([]presentation.CouplePreview, error)
	FollowersToNotify(copuleID entity.ID, skip int) ([]entity.ID, error)
	SuggestedAccounts(exempted []entity.ID, country string) ([]presentation.CouplePreview, error)
	IsBlocked(coupleName string, userID entity.ID) (bool, error)
}

//Repository all couple methods
type Repository interface {
	Writer
	Reader
}

//Couple usecase
type UseCase interface {
	CreateCouple(userId, partnerrId, coupleName, country, state string) (entity.ID, error)
	UpdateCouple(coupleID entity.ID, update entity.UpdateCouple) error
	GetCouple(coupleName string, userID entity.ID) (entity.Couple, error)
	GetCouplePosts(coupleName string, skip int) (entity.Couple, error)
	GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error)
	GetFollowers(coupleName string, skip int) ([]entity.ID, error)
	SearchCouples(query string, userID entity.ID) ([]presentation.CouplePreview, error)
	NewFollower(userID, coupleID entity.ID) error
	RemoveFollower(coupleID, userID entity.ID) error
	UpdateCoupleProfilePic(fileName string, coupleID entity.ID) error
	UpdateCoupleCoverPic(fileName string, coupleID entity.ID) error
	ChangeCoupleName(coupleID entity.ID, coupleName string) error
	BreakUp(coupleID entity.ID) error
	MakeUp(coupleID entity.ID) error
	WhereACouple(userID, partnerID entity.ID) (entity.ID, error)
	ListCouple(coupleIDs []entity.ID, userID entity.ID) ([]presentation.CouplePreview, error)
	AddPost(coupleID entity.ID, postID string) error
	RemovePost(coupleID entity.ID, postID string) error
	UpdateStatus(coupleID entity.ID, married bool) error
	FollowersToNotify(copuleID entity.ID, skip int) ([]entity.ID, error)
	GetSuggestedAccounts(exempted []entity.ID, country string) ([]presentation.CouplePreview, error)
	BlockUser(coupleID, userID entity.ID) error
	IsBlocked(couleName string, userID entity.ID) (bool, error)
}
