package ioc

import (
	"context"
	"time"
	"webook/pkg/logger"
	"webook/tag/repository"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"
)

func InitRepository(dao dao.TagDao, cache cache.TagCache, l logger.LoggerV1) repository.TagRepository {
	repo := repository.NewCachedTagRepository(dao, cache, l)
	go func() {
		// 执行缓存预加载
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// 也可以同步执行。但是在一些场景下，同步执行会占用非常长得时间，所以可以考虑异步执行
		repo.PreloadUserTags(ctx)
	}()
	return repo
}
