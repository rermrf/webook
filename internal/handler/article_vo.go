package handler

// vo: view object, 对标前端的

type ArticleVO struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	// Author 要从用户来
	AuthorId   int64  `json:"author_id"`
	AuthorName string `json:"author_name"`
	// 状态可以是前端来处理，也可以是后端来处理
	// 0 -> unknown -> 未知状态
	// 1 -> unpublish -> 未发表状态
	// 2 -> publish -> 已发表
	// 3 -> private -> 私密
	Status uint8 `json:"status"`
	// 计数
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`

	// 我个人有没有收藏，有没有点赞
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`

	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ListReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type WithdrawReq struct {
	Id int64
}

type LikeReq struct {
	// 点赞和取消点赞都复用这个接口
	Like bool  `json:"like"`
	Id   int64 `json:"id"`
}
