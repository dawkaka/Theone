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

func (s *Service) CreateCouple(userId, partnerId string) error {
	initiated, err := entity.StringToID(partnerId)
	if err != nil {
		return err
	}
	accepted, err := entity.StringToID(userId)
	if err != nil {
		return err
	}

	couple := entity.Couple{
		Iniated:        initiated,
		Accepted:       accepted,
		IniatedAt:      time.Time{},
		AcceptedAt:     time.Now(),
		CoupleName:     "",
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

func (s *Service) UpdateCouple(couple entity.Couple) error {

	return s.repo.Update(couple)
}
