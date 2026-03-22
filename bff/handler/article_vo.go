package handler

// vo: view object, 对标前端的

type ArticleVO struct {
	Id       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Content  string `json:"content"`
	// Author 要从用户来
	AuthorId   int64  `json:"author_id"`
	AuthorName      string `json:"author_name"`
	CoverUrl        string `json:"cover_url,omitempty"`
	AuthorAvatarUrl string `json:"author_avatar_url,omitempty"`
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

	// 是否关注了作者
	AuthorFollowed bool `json:"author_followed"`

	Ctime string `json:"ctime"`
	Utime string `json:"utime"`
}

type ArticleReq struct {
	Id      int64   `json:"id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Tags    []int64 `json:"tags"`
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

type CollectReq struct {
	// 点赞和取消点赞都复用这个接口
	Collect bool  `json:"collect"`
	Id      int64 `json:"id"`
}

type RewardReq struct {
	Id  int64 `json:"id"`
	Amt int64 `json:"amt"`
}

type InteractiveReq struct {
	Id int64 `form:"id"`
}

type InteractiveResp struct {
	Biz        string
	BizId      int64
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Liked      bool
	Collected  bool
}

type GetCommentReq struct {
	Id int64 `form:"id"`
	// 从最小评论的id起
	MinId int64 `form:"min_id"`
	Limit int64 `form:"limit"`
}
type GetCommentCntReq struct {
	Id int64 `form:"id"`
}

type GetCommentCntResp struct {
	Cnt int64 `json:"cnt"`
}

type Comment struct {
	Id int64 `json:"id"`
	//BizId   int64  `json:"biz_id"`
	Content  string `json:"content"`
	Uid      int64  `json:"uid"`
	UserName      string `json:"user_name"`
	UserAvatarUrl string `json:"user_avatar_url,omitempty"`
	ParentId int64  `json:"parent_id"`
	RootId   int64  `json:"root_id"`
	Ctime    int64  `json:"ctime"`
}

type GetCommentResp struct {
	Comments []Comment
}

type CreateCommentReq struct {
	Id       int64  `json:"id"`
	Content  string `json:"content"`
	ParentId int64  `json:"parent_id"`
	RootId   int64  `json:"root_id"`
}

type GetPubListReq struct {
	Offset int32 `form:"offset"`
	Limit  int32 `form:"limit"`
	//startTime int64 `form:"start_time"`
}

type DeleteReq struct {
	Id int64 `json:"id"`
}

type GetMoreRepliesReq struct {
	MinId int64 `form:"min_id"`
	Limit int64 `form:"limit"`
}

type ListLikedReq struct {
	Biz    string `form:"biz"`
	Offset int64  `form:"offset"`
	Limit  int64  `form:"limit"`
}

type ListCollectedReq struct {
	Biz    string `form:"biz"`
	Offset int64  `form:"offset"`
	Limit  int64  `form:"limit"`
}
