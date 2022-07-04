package entity

import (
	"time"

	"github.com/google/uuid"
)

//ID entity ID
type ID = uuid.UUID

type Comment struct {
	UserID    ID        `json:"id"`
	Comment   string    `json:"comment"`
	Likes     []ID      `json:"likes"`
	CreatedAt time.Time `json:"created_at"`
}

//NewID create a new entity ID
func NewID() ID {
	return ID(uuid.New())
}

//StringToID convert a string to an entity ID
func StringToID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	return ID(id), err
}
