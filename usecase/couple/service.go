package couple

import (
	"time"

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

func (s *Service) GetCouple(coupleName string) (entity.Couple, error) {
	return s.repo.Get(coupleName)
}

func (s *Service) GetCouplePosts(coupleName string, skip int) ([]entity.Post, error) {
	return s.repo.GetCouplePosts(coupleName, skip)
}

func (s *Service) GetCoupleVideos(coupleName string, skip int) ([]entity.Video, error) {
	return s.repo.GetCoupleVideos(coupleName, skip)
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
		Initiated:      initiated,
		Accepted:       accepted,
		AcceptedAt:     time.Now(),
		CoupleName:     coupleName,
		Married:        false,
		Verified:       false,
		ProfilePicture: "defaultProfile.jpg",
		CoverPicture:   "defaultCover.jpg",
		Bio:            "-",
		Followers:      []primitive.ObjectID{},
		FollowersCount: 0,
		PostCount:      0,
		Status:         "In a relationship",
	}
	return s.repo.Create(couple)
}

func (s *Service) GetFollowers(coupleName string, skip int) ([]entity.Follower, error) {
	return s.repo.Followers(coupleName, skip)
}

func (s *Service) NewFollower(userID, coupleID entity.ID) error {
	return s.repo.Follower(userID, coupleID)
}

func (s *Service) RemoveFollower(userID, coupleID entity.ID) error {
	return s.repo.Unfollow(userID, coupleID)
}

func (s *Service) UpdateCouple(couple entity.Couple) error {

	return s.repo.Update(couple)
}

func (s *Service) UpdateCoupleProfilePic(fileName string, coupleID entity.ID) error {
	return s.repo.UpdateProfilePic(fileName, coupleID)
}

func (s *Service) UpdateCoupleCoverPic(fileName string, coupleID entity.ID) error {
	return s.repo.UpdateCoverPic(fileName, coupleID)
}
