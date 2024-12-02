package events

type SyncUserEvent struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}
