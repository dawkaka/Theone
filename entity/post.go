package entity

import (
	"strings"
	"time"

	"github.com/dawkaka/theone/pkg/validator"
)

//Post data
type Post struct {
	ID            ID        `json:"id" bson:"_id"`
	PostID        string    `json:"post_id" bson:"post_id"`
	CoupleID      ID        `json:"couple_id" bson:"couple_id"`
	InitiatedID   ID        `json:"initiated_id" bson:"initiated_id"`
	AcceptedID    ID        `json:"accepted_id" bson:"accepted_id"`
	PostedBy      ID        `json:"posted_by" bson:"posted_by"`
	FileName      string    `json:"file_name" bson:"file_name"`
	Caption       string    `json:"caption"`
	Mentioned     []string  `json:"mentioned"`
	Likes         []ID      `json:"likes"`
	LikesCount    int64     `json:"likes_count" bson:"likes_count"`
	Comments      []Comment `json:"comment"`
	CommentsCount int64     `json:"comments_count" bson:"comments_count"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	Type          string    `json:"type"`
}

func (p *Post) Sanitize() {
	p.Caption = strings.TrimSpace(p.Caption)
}

func (p Post) Validate() []error {
	errs := []error{}
	if !validator.IsCaption(p.Caption) {
		errs = append(errs, ErrInvalidCaption)
	}
	return errs
}
