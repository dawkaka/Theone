package entity

import "time"

type CoupleMessage struct {
	From     ID        `json:"from"`
	To       ID        `json:"to"`
	Date     time.Time `json:"date"`
	Type     string    `json:"type"`
	Message  string    `json:"message"`
	Caption  string    `json:"caption"`
	Recieved bool      `json:"recieved"`
	CoupleID ID        `json:"couple_id" bson:"couple_id"`
}

type UserCoupleMessage struct {
	From     ID        `json:"from"`
	To       ID        `json:"to"`
	Date     time.Time `json:"date"`
	Type     string    `json:"type"`
	Message  string    `json:"message"`
	Caption  string    `json:"caption"`
	Recieved bool      `json:"recieved"`
	AFrom    ID        `json:"afrom" bson:"afrom"`
}
