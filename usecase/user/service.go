package user

import (
	"strings"
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
func (s *Service) CreateUser(email, password, firstName, lastName string) (entity.ID, error) {
	e := entity.NewUser(email, firstName, lastName, time.Date(1999, time.March, 16, 0, 0, 0, 0, time.Now().Location()))

	return s.repo.Create(e)
}

//GetUser Get an user
func (s *Service) GetUser(id entity.ID) (*entity.User, error) {
	return s.repo.Get(id)
}

//SearchUsers Search users
func (s *Service) SearchUsers(query string) ([]*entity.User, error) {
	return s.repo.Search(strings.ToLower(query))
}

//ListUsers List users
func (s *Service) ListUsers(users []entity.ID) ([]*entity.User, error) {
	return s.repo.List(users)
}

//DeleteUser Delete an user
func (s *Service) DeleteUser(id entity.ID) error {
	u, err := s.GetUser(id)
	if u == nil {
		return entity.ErrNotFound
	}
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

//UpdateUser Update an user
func (s *Service) UpdateUser(e *entity.User) error {
	err := e.Validate()
	if err != nil {
		return entity.ErrInvalidEntity
	}
	e.UpdatedAt = time.Now()
	return s.repo.Update(e)
}
