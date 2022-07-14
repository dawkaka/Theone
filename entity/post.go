package entity

import "time"

//Post data
type Post struct {
	ID            ID        `json:"id" bson:"_id"`
	PostID        string    `json:"post_id" bson:"post_id"`
	CoupleID      ID        `json:"couple_id" bson:"couple_id"`
	FileName      string    `json:"file_name" bson:"file_name"`
	Caption       string    `json:"caption"`
	Likes         []ID      `json:"likes"`
	LikesCount    int64     `json:"likes_count" bson:"likes_count"`
	Comments      []Comment `json:"comment"`
	CommentsCount int64     `json:"comments_count" bson:"comments_count"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	Type          string    `json:"type"`
}
