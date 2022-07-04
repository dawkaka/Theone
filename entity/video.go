package entity

import "time"

//Video data
type Video struct {
	ID         ID            `json:"id"`
	CoupleID   ID            `json:"couple_id"`
	Caption    string        `json:"caption"`
	Likes      []ID          `json:"likes"`
	CoverImage string        `json:"cover_image"`
	Comments   []Comment     `json:"comment"`
	ViewCount  uint64        `json:"view_count"`
	CreatedAt  time.Time     `json:"created_at"`
	Type       string        `json:"type"`
	FileName   string        `json:"file_name"`
	Duration   time.Duration `json:"duration"`
	Size       uint64        `json:"size"`
	Height     uint16        `json:"height"`
	Width      uint16        `json:"width"`
}
