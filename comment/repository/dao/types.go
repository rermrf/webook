package dao

import (
	"context"
	"database/sql"
)

type CommentDao interface {
	Inser(ctx context.Context, u Comment) error
	// FindByBiz 只查找一级评论
	FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error)
	// FindCommentList 当 comment 的 id 为 0 ，则获取一级评论，如果不为0获取对应的评论及其评论的所有回复
	FindCommentList(ctx context.Context, c Comment) ([]Comment, error)
	FindRepliesByPid(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	// Delete 删除本节点及其对于的子节点
	Delete(ctx context.Context, c Comment) error
	FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, id int64, limit int64) ([]Comment, error)
}

type Comment struct {
	Id int64 `gorm:"primaryKey" json:"id"`
	// 发表评论的用户
	Uid int64 `gorm:"index" json:"uid"`
	// 发表评论的业务类型
	Biz string `gorm:"index:biz_type_id" json:"biz"`
	// 对应业务的 ID
	BizId int64 `gorm:"index:biz_type_id" json:"biz_id"`
	// 根评论
	RootId sql.NullInt64 `gorm:"index" json:"root_id"`
	// 父级评论
	PID sql.NullInt64 `gorm:"index" json:"pid"`
	// 外键 用于联级删除
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`
	// 评论内容
	Content string `gorm:"type:text" json:"content"`
	Ctime   int64  `json:"ctime"`
	Utime   int64  `json:"utime"`
}

func (*Comment) TableName() string {
	return "comments"
}
