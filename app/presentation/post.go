package presentation

import (
	"time"

	"github.com/dawkaka/theone/entity"
)

type Post struct {
	ID            entity.ID             `json:"id"`
	CoupleName    string                `json:"couple_name"`
	CreatedAt     time.Time             `json:"created_at"`
	Caption       string                `json:"caption"`
	LikesCount    int64                 `json:"likes_count"`
	CommentsCount int64                 `json:"comments_count"`
	Files         []entity.PostMetadata `json:"files"`
}

type Comment struct {
	ID         entity.ID `json:"id" bson:"_id"`
	Comment    string    `json:"comment" bson:"comment"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	LikesCount int       `json:"likes_count" bson:"likes_count"`
	UserName   string    `json:"user_name" bson:"user_name"`
	HasPartner bool      `json:"has_partner" bson:"has_partner"`
}