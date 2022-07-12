package user

import (
	"time"

	"github.com/dawkaka/theone/entity"
)

//Service  interface
type Service struct {
	repo Repository
}

//NewService create new use case
func NewService(r Repository) *Service {
	return &Service{
		repo: r,
	}
}

//CreateUser Create an user
func (s *Service) CreateUser(email, password, firstName, lastName, userName string, dateOfBirth time.Time) error {
	e := entity.NewUser(email, password, firstName, lastName, userName, dateOfBirth)
	return s.repo.Create(e)
}

func (s *Service) CreateRequest(from, to entity.ID) error {
	return s.repo.Request(from, to)
}

//GetUser Get an user
func (s *Service) GetUser(userName string) (*entity.User, error) {
	return s.repo.Get(userName)
}

//SearchUsers Search users
func (s *Service) SearchUsers(query string) ([]*entity.User, error) {
	return s.repo.Search(query)
}

//ListUsers List users
func (s *Service) ListUsers(users []entity.ID) ([]entity.User, error) {
	return s.repo.List(users)
}

//DeleteUser Delete an user
func (s *Service) DeleteUser(id entity.ID) error {
	return s.repo.Delete(id)
}

//UpdateUser Update an user
func (s *Service) UpdateUser(e *entity.User) error {
	e.UpdatedAt = time.Now()
	return s.repo.Update(e)
}

func (s *Service) ConfirmCouple(userID, partnerID string) (bool, error) {
	user, err := entity.StringToID(userID)
	if err != nil {
		return false, err
	}
	parnter, err := entity.StringToID(partnerID)

	if err != nil {
		return false, err
	}
	return s.repo.ConfirmCouple(user, parnter), nil
}

func (s *Service) NotifyUser(user string, notif any) error {
	return s.repo.Notify(user, notif)
}
