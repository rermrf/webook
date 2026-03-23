package dao

import "gorm.io/gorm"

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&CreditAccount{},
		&CreditFlow{},
		&DailyLimit{},
		&CreditRule{},
		&CreditReward{},
		&EpayOrder{},
	)
}

// InitDefaultRules 初始化默认积分规则
func InitDefaultRules(db *gorm.DB) error {
	rules := []CreditRule{
		{Biz: "read", CreditAmt: 1, DailyLimit: 50, Description: "阅读文章+1积分,每日上限50", Enabled: true},
		{Biz: "like", CreditAmt: 2, DailyLimit: 30, Description: "点赞文章+2积分,每日上限60", Enabled: true},
		{Biz: "collect", CreditAmt: 3, DailyLimit: 20, Description: "收藏文章+3积分,每日上限60", Enabled: true},
		{Biz: "comment", CreditAmt: 5, DailyLimit: 20, Description: "评论文章+5积分,每日上限100", Enabled: true},
	}

	for _, rule := range rules {
		// 只在规则不存在时插入
		var count int64
		db.Model(&CreditRule{}).Where("biz = ?", rule.Biz).Count(&count)
		if count == 0 {
			if err := db.Create(&rule).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
