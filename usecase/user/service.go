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

//GetUser Get an user
func (s *Service) GetUser(userName string) (*entity.User, error) {
	return s.repo.Get(userName)
}

//SearchUsers Search users
func (s *Service) SearchUsers(query string) ([]*entity.User, error) {
	return s.repo.Search(query)
}

//ListUsers List users
func (s *Service) ListUsers(users []entity.ID) ([]*entity.User, error) {
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
