package domain

type BrowseHistory struct {
	Id         int64
	UserId     int64
	Biz        string
	BizId      int64
	BizTitle   string
	AuthorName string
	Ctime      int64
	Utime      int64
}
