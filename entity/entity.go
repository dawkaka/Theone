package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//ID entity ID
type ID = primitive.ObjectID
type Comment struct {
	UserID    string    `json:"user_id"`
	Comment   string    `json:"comment"`
	Likes     []ID      `json:"likes"`
	CreatedAt time.Time `json:"created_at"`
}

//StringToID convert a string to an entity ID
func StringToID(s string) (ID, error) {
	return primitive.ObjectIDFromHex(s)
}

type Pagination struct {
	Next int  `json:"next"`
	End  bool `json:"end"`
}

var (
	Limit  = 30
	LimitP = 15
)
