package dao

import (
	"context"
	"database/sql"
)

type CommentDao interface {
	Insert(ctx context.Context, u Comment) error
	// FindByBiz 只查找一级评论
	FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error)
	// FindCommentList 当 comment 的 id 为 0 ，则获取一级评论，如果不为0获取对应的评论及其评论的所有回复
	FindCommentList(ctx context.Context, c Comment) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	// Delete 删除本节点及其对于的子节点
	Delete(ctx context.Context, c Comment) error
	FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, maxId int64, limit int64) ([]Comment, error)
}

// Comment 总结：所有的索引设计，都是针对 WHERE、ORDER BY、SELECT xxx 来进行的
// 如果有 JOIN，那么还要考虑 ON
// 永远考虑最频繁的查询
// 在没有遇到更新、查询性能瓶颈之前，不需要太过于担忧维护索引的开销
// 有一些时候，随着业务发展，有一些索引用不上了，要及时删除
type Comment struct {
	Id int64 `gorm:"primaryKey" json:"id"`
	// 发表评论的用户
	Uid int64 `gorm:"index" json:"uid"`
	// 发表评论的业务类型
	Biz string `gorm:"index:biz_type_id" json:"biz"`
	// 对应业务的 ID
	BizId int64 `gorm:"index:biz_type_id" json:"biz_id"`
	// 根评论
	RootId sql.NullInt64 `gorm:"index:root_id_ctime" json:"root_id"`
	// 父级评论
	PID sql.NullInt64 `gorm:"column:pid;index" json:"pid"`
	// 外键 用于联级删除
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	// 评论内容
	Content string `gorm:"type:text" json:"content"`
	Ctime   int64  `gorm:"index:root_id_ctime" json:"ctime"`
	Utime   int64  `json:"utime"`
}

func (*Comment) TableName() string {
	return "comments"
}
