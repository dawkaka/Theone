package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//ID entity ID
type ID = primitive.ObjectID

type Comment struct {
	ID        ID        `json:"id" bson:"_id,omitempty"`
	UserID    ID        `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Likes     []ID      `json:"likes,omitempty" bson:"likes, omitempty"`
}

//StringToID convert a string to an entity ID
func StringToID(s string) (ID, error) {
	return primitive.ObjectIDFromHex(s)
}

type Pagination struct {
	Next int  `json:"next"`
	End  bool `json:"end"`
}

const (
	Limit          = 10
	LimitP         = 6
	EIGHTEEN_YEARS = 18
)
const (
	NO_REQUEST int8 = iota
	SENT_REQUEST
	RECIEVED_REQUEST
)

type PhotoMetaData struct {
	Name string `json:"file_name" bson:"file_name"`
	Size int64  `json:"file_size" bson:"file_size"`
}
