package couple

import (
	"time"

	"github.com/dawkaka/theone/app/presentation"
	"github.com/dawkaka/theone/entity"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

func (s *Service) GetCouple(coupleName string, userID entity.ID) (entity.Couple, error) {
	return s.repo.Get(coupleName, userID)
}

func (s *Service) GetCouplePosts(coupleName string, skip int) (entity.Couple, error) {
	return s.repo.GetCouplePosts(coupleName, skip)
}

func (s *Service) GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error) {
	return s.repo.GetCoupleVideos(coupleName, skip)
}

func (s *Service) ListCouple(IDs []entity.ID, userID entity.ID) ([]presentation.CouplePreview, error) {
	return s.repo.List(IDs, userID)
}

func (s *Service) CreateCouple(userId, partnerId, coupleName string) (entity.ID, error) {
	initiated, err := entity.StringToID(partnerId)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	accepted, err := entity.StringToID(userId)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	couple := entity.Couple{
		ID:             primitive.NewObjectID(),
		Initiated:      initiated,
		Accepted:       accepted,
		AcceptedAt:     time.Now(),
		CoupleName:     coupleName,
		DateCommenced:  time.Now(),
		Married:        false,
		Verified:       false,
		ProfilePicture: "defaultProfile.jpg",
		CoverPicture:   "defaultCover.jpg",
		Bio:            "-",
		Followers:      []primitive.ObjectID{},
		FollowersCount: 0,
		PostCount:      0,
		Posts:          []string{},
	}
	return s.repo.Create(couple)
}

func (s *Service) GetFollowers(coupleName string, skip int) ([]entity.ID, error) {
	return s.repo.Followers(coupleName, skip)
}

func (s *Service) NewFollower(userID, coupleID entity.ID) error {
	return s.repo.Follower(userID, coupleID)
}

func (s *Service) RemoveFollower(coupleID, userID entity.ID) error {
	return s.repo.Unfollow(coupleID, userID)
}

func (s *Service) UpdateCouple(coupleID entity.ID, update entity.UpdateCouple) error {
	update.UpdatedAt = time.Now()
	return s.repo.Update(coupleID, update)
}

func (s *Service) UpdateCoupleProfilePic(fileName string, coupleID entity.ID) error {
	return s.repo.UpdateProfilePic(fileName, coupleID)
}

func (s *Service) UpdateCoupleCoverPic(fileName string, coupleID entity.ID) error {
	return s.repo.UpdateCoverPic(fileName, coupleID)
}

func (s *Service) ChangeCoupleName(coupleID entity.ID, coupleName string) error {
	return s.repo.ChangeName(coupleID, coupleName)
}

func (s *Service) BreakUp(coupleID entity.ID) error {
	return s.repo.BreakUp(coupleID)
}

func (s *Service) MakeUp(coupleID entity.ID) error {
	return s.repo.MakeUp(coupleID)
}

func (s *Service) WhereACouple(userID, partnerID entity.ID) (entity.ID, error) {
	return s.repo.Dated(userID, partnerID)
}

func (s *Service) AddPost(coupleID entity.ID, postID string) error {
	return s.repo.AddPost(coupleID, postID)
}

func (s *Service) RemovePost(coupleID entity.ID, postID string) error {
	return s.repo.RemovePost(coupleID, postID)
}

func (s *Service) SearchCouples(query string) ([]presentation.CouplePreview, error) {
	return s.repo.Search(query)
}

func (s *Service) UpdateStatus(coupleID entity.ID, married bool) error {
	return s.repo.UpdateStatus(coupleID, married)
}

func (s *Service) FollowersToNotify(coupleID entity.ID, skip int) ([]entity.ID, error) {
	return s.repo.FollowersToNotify(coupleID, skip)
}
