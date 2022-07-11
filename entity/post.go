package entity

import "time"

//Post data
type Post struct {
	ID        ID        `json:"id"`
	CoupleID  ID        `json:"couple_id"`
	Caption   string    `json:"caption"`
	Likes     []ID      `json:"likes"`
	Comments  []Comment `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"`
}
