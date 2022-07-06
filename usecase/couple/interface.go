package couple

import "github.com/dawkaka/theone/entity"

//Writer couple writer methods
type Writer interface {
	Create(couple entity.Couple) error
	Update(couple entity.Couple) error
}

//Reader couple reader methods
type Reader interface {
	Get(coupleName string) (entity.Couple, error)
}

//Repository all couple methods
type Repository interface {
	Writer
	Reader
}

//Couple usecase
type UseCase interface {
	CreateCouple(couple entity.Couple) error
	UpdateCouple(couple entity.Couple) error
	GetCouple(coupleName string) (entity.Couple, error)
}
