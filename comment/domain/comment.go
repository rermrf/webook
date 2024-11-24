package domain

import "time"

type Comment struct {
	Id int64 `json:"id"`
	// 评论者
	Commentator User `json:"user"`
	// 评论对象
	// 数据里面
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
	// 评论内容
	Content string `json:"content"`
	// 根评论
	RootComment *Comment `json:"root_comment"`
	// 父评论
	ParentComment *Comment  `json:"parent_comment"`
	Children      []Comment `json:"children"`
	Ctime         time.Time `json:"ctime"`
	Utime         time.Time `json:"utime"`
}

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
