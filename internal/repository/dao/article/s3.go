package article

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
	"webook/internal/domain"
)

type S3DAO struct {
	oss *s3.S3
	// 通过组合 GORMArticleDAO 来简化操作
	// 操作制作库是一样的
	GormArticleDao
	bucket *string
}

func NewOssDAO(oss *s3.S3, db *gorm.DB) ArticleDao {
	return &S3DAO{
		oss:    oss,
		bucket: aws.String("webook-1258698140"),
		GormArticleDao: GormArticleDao{
			db: db,
		},
	}
}

// Sync 制作库存储所有的数据，线上库只存储参与sql运算的数据，content 存储到oss
func (dao *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 保存制作库
	// 保存线上库，并且把 content 上传到oss
	var (
		id = art.Id
	)
	// 制作库流量不大，并发不高，保存到数据库存储
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDao := NewGormArticleDao(tx)
		// 制作库
		if id > 0 {
			err = txDao.UpdateById(ctx, art)
		} else {
			id, err = txDao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		now := time.Now().UnixMilli()
		art.Id = id
		publishArt := PublishArticle(art)
		publishArt.Utime = now
		publishArt.Ctime = now
		// 线上库不会保存这个 Content，要准备上传到 OSS 里面
		publishArt.Content = ""
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上 MYSQL 里面写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  art.Title,
				"utime":  now,
				"status": art.Status,
			}),
		}).Create(&publishArt).Error
	})
	// 说明保存到数据库的时候失败了
	if err != nil {
		return 0, err
	}
	// 保存到 OSS
	// 有可能保存到数据库成功，但是保存到 oss 失败
	// 监控，重试，补偿机制
	_, err = dao.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      dao.bucket,
		Key:         aws.String(strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: aws.String("text/plain;charset=utf-8"),
	})
	return id, err
}

func (dao *S3DAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("非法修改别人的文章状态")
		}

		res = tx.Model(&PublishArticle{}).
			Where("id = ? AND author_id = ?", id, author).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return errors.New("非法修改别人的文章状态")
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 隐藏的话删除oss中的文件
	if status == domain.ArticleStatusPrivate.ToUint8() {
		_, err = dao.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: dao.bucket,
			Key:    aws.String(strconv.FormatInt(id, 10)),
		})
	}
	return err
}
