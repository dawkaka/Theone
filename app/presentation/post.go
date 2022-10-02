package presentation

import (
	"time"

	"github.com/dawkaka/theone/entity"
)

type Post struct {
	ID             entity.ID             `json:"id" bson:"_id"`
	CoupleName     string                `json:"couple_name"`
	Married        bool                  `json:"married"`
	Verified       bool                  `json:"verified"`
	ProfilePicture string                `json:"profile_picture"`
	CreatedAt      time.Time             `json:"created_at" bson:"created_at"`
	Caption        string                `json:"caption"`
	LikesCount     int64                 `json:"likes_count" bson:"likes_count"`
	CommentsCount  int64                 `json:"comments_count" bson:"comments_count"`
	Files          []entity.PostMetadata `json:"files" bson:"files"`
	IsThisCouple   bool                  `json:"is_this_couple"`
	Location       string                `json:"location"`
	HasLiked       bool                  `json:"has_liked"`
}

type Comment struct {
	entity.Comment `bson:"inline"`
	UserName       string `json:"user_name" bson:"user_name"`
	HasPartner     bool   `json:"has_partner" bson:"has_partner"`
	LikesCount     int    `json:"likes_count" bson:"likes_count"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
	HasLiked       bool   `json:"has_liked" bson:"has_liked"`
}
