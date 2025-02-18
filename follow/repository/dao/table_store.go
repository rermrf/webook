package dao

import (
	"context"
	"fmt"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"os"
	"time"
)

const FollowTableName = "follow"

type TableStoreFollowDao struct {
	client *tablestore.TableStoreClient
}

func NewTableStoreFollowDao() FollowDao {
	endpoint := os.Getenv("TS_ENDPOINT")
	accessId := os.Getenv("TS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("TS_ACCESS_KEY_SECRET")
	instanceName := os.Getenv("TS_INSTANCE_NAME")
	client := tablestore.NewClient(endpoint, accessId, accessKeySecret, instanceName)
	return &TableStoreFollowDao{client: client}
}

func (t *TableStoreFollowDao) FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error) {
	req := &tablestore.SQLQueryRequest{
		// 有 sql 注入的隐患
		// 在需要前端传入参数拼接 sql 语句，要小心 sql 注入
		// "select id, follower, followee form %s where follower = %s AND status = %d OFFSET %d LIMIT %d"
		// "select id, follower, followee form 'follow' where follower = 1 or 1 = 1 AND status = 1 OFFSET 0 LIMIT 10"
		Query: fmt.Sprintf("select id, follower, followee form %s where follower = %d AND status = %d OFFSET %d LIMIT %d", FollowTableName, follower, FollowRelationStatusActive, offset, limit),
	}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return nil, err
	}
	resSet := resp.ResultSet
	followRelations := make([]FollowRelation, 0, limit)
	for resSet.HasNext() {
		row := resSet.Next()
		followRelation := FollowRelation{}
		followRelation.Id, _ = row.GetInt64ByName("id")
		followRelation.Follower, _ = row.GetInt64ByName("follower")
		followRelation.Followee, _ = row.GetInt64ByName("followee")
		followRelations = append(followRelations, followRelation)
	}
	return followRelations, nil
}

func (t *TableStoreFollowDao) FollowerRelationList(ctx context.Context, followee, offset, limit int64) ([]FollowRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TableStoreFollowDao) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TableStoreFollowDao) CreateFollowRelation(ctx context.Context, c FollowRelation) error {
	now := time.Now().UnixMilli()
	// UpdateRowRequest + RowExistenceExpectation_IGNORE
	// 可以实现一个 insert or update 的语义
	// 单纯的使用 update 或 put 不能达成这个效果
	req := new(tablestore.UpdateRowRequest)
	pk := &tablestore.PrimaryKey{}
	pk.AddPrimaryKeyColumn("follower", c.Follower)
	pk.AddPrimaryKeyColumn("followee", c.Followee)
	change := &tablestore.UpdateRowChange{
		TableName: FollowTableName,
		// 有个小的问题，这边其实不用 id，直接用 follower 和 followee 构成主键
		// 如果要用 ID，你可以用自增主键
		PrimaryKey: pk,
	}
	// 设置
	change.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	// 只能用 int64
	change.PutColumn("status", int64(c.Status))
	change.PutColumn("ctime", now)
	change.PutColumn("utime", now)
	req.UpdateRowChange = change
	_, err := t.client.UpdateRow(req)
	return err
}

func (t *TableStoreFollowDao) UpdateStatus(ctx context.Context, follower int64, followee int64, status uint8) error {
	cond := tablestore.NewCompositeColumnCondition(tablestore.LO_AND)
	// 更新条件，对标 where
	// 多个 Filter 是 AND 条件连在一起
	cond.AddFilter(tablestore.NewSingleColumnCondition("follower", tablestore.CT_EQUAL, follower))
	cond.AddFilter(tablestore.NewSingleColumnCondition("followee", tablestore.CT_EQUAL, followee))
	change := new(tablestore.UpdateRowChange)
	change.TableName = FollowTableName
	// 我预期这一行数据是存在的
	// 不在就报错
	change.SetCondition(tablestore.RowExistenceExpectation_EXPECT_EXIST)
	change.SetColumnCondition(cond)
	change.PutColumn("status", int64(status))
	_, err := t.client.UpdateRow(&tablestore.UpdateRowRequest{
		UpdateRowChange: change,
	})
	return err
}

func (t *TableStoreFollowDao) CntFollower(ctx context.Context, uid int64) (int64, error) {
	req := &tablestore.SQLQueryRequest{
		Query: fmt.Sprintf("select count(follower) as cnt from %s where follower = %d AND status = %d", FollowTableName, uid, FollowRelationStatusActive),
	}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return 0, err
	}
	res := resp.ResultSet
	if res.HasNext() {
		row := res.Next()
		return row.GetInt64ByName("cnt")
	}
	return 0, nil
}

func (t *TableStoreFollowDao) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	req := &tablestore.SQLQueryRequest{
		Query: fmt.Sprintf("select count(followee) as cnt from %s where followee = %d AND status = %d", FollowTableName, uid, FollowRelationStatusActive),
	}
	resp, err := t.client.SQLQuery(req)
	if err != nil {
		return 0, err
	}
	res := resp.ResultSet
	if res.HasNext() {
		row := res.Next()
		return row.GetInt64ByName("cnt")
	}
	return 0, nil
}
