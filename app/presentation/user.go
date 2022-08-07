package presentation

type UserProfile struct {
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	UserName       string    `json:"user_name"`
	ProfilePicture string    `json:"profile_picture"`
	Bio            string    `json:"bio"`
	FollowingCount uint64    `json:"following_count"`
	ShowPictures   [6]string `json:"show_pictures"`
	HasPartener    bool      `json:"has_partner"`
}

type UserPreview struct {
	FirstName      string `json:"first_name" bson:"first_name"`
	LastName       string `json:"Last_naem" bson:"last_name"`
	UserName       string `json:"user_naame" bson:"user_name"`
	HasPartener    bool   `json:"has_partner" bson:"hast_partner"`
	ProfilePicture string `json:"profile_picture" bson:"profile_picture"`
}
