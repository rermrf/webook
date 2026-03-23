package events

// ReadEvent 阅读文章事件
type ReadEvent struct {
	Uid   int64  `json:"uid"`
	Aid   int64  `json:"aid"`
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
}

// LikeEvent 点赞事件
type LikeEvent struct {
	Uid    int64  `json:"uid"`
	Biz    string `json:"biz"`
	BizId  int64  `json:"biz_id"`
	Action string `json:"action"` // like or cancel
}

// CollectEvent 收藏事件
type CollectEvent struct {
	Uid    int64  `json:"uid"`
	Biz    string `json:"biz"`
	BizId  int64  `json:"biz_id"`
	Action string `json:"action"` // collect or cancel
}
