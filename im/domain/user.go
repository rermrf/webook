package domain

type User struct {
	Nickname string `json:"nickname"`
	UserId   string `json:"userId"`
	FaceURL  string `json:"faceUrl"`
}
