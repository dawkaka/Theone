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
	Limit          = 30
	LimitP         = 15
	EIGHTEEN_YEARS = 18
)

type PhotoMetaData struct {
	Name string `json:"file_name" bson:"file_name"`
	Size int64  `json:"file_size" bson:"file_size"`
}
