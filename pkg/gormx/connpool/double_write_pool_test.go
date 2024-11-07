package connpool

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestConnPool(t *testing.T) {
	webook, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local"))
	require.NoError(t, err)
	err = webook.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	intr, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13306)/webook_intr?charset=utf8mb4&parseTime=True&loc=Local"))
	require.NoError(t, err)
	err = intr.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DoubleWritePool{
			src:     webook.ConnPool,
			dst:     intr.ConnPool,
			pattern: *atomic.NewString(PatternSrcFirst),
		},
	}))
	require.NoError(t, err)
	t.Log(db)
	err = db.Create(&Interactive{
		Biz:   "test",
		BizId: 1230008,
	}).Error
	require.NoError(t, err)

	err = db.Transaction(func(tx *gorm.DB) error {
		err1 := tx.Create(&Interactive{
			Biz:   "test_tx",
			BizId: 12300014,
		}).Error
		return err1
	})
	require.NoError(t, err)

	err = db.Model(&Interactive{}).Where("id = 236").Updates(map[string]interface{}{
		"biz_id": 7890,
	}).Error
	require.NoError(t, err)
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_id_type"`
	Biz        string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Ctime      int64
	Utime      int64
}
