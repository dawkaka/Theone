package entity

import "time"

//Video data
type Video struct {
	ID            ID            `json:"id"`
	VideoID       string        `json:"video_id" bson:"video_id"`
	CoupleID      ID            `json:"couple_id" bson:"couple_id"`
	Caption       string        `json:"caption"`
	Likes         []ID          `json:"likes"`
	LikesCount    int64         `json:"likes_count" bson:"likes_count"`
	Thumbnail     string        `json:"cover_image" bson:"thumbnail"`
	Comments      []Comment     `json:"comment"`
	CommentsCount int64         `json:"comments_count" bson:"comments_count"`
	ViewCount     uint64        `json:"view_count" bson:"view_count"`
	CreatedAt     time.Time     `json:"created_at" bson:"created_at"`
	Type          string        `json:"type"`
	FileName      string        `json:"file_name" bson:"file_name"`
	Duration      time.Duration `json:"duration"`
	Size          uint64        `json:"size"`
	Height        uint16        `json:"height"`
	Width         uint16        `json:"width"`
}
