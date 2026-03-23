package ioc

import (
	"webook/pkg/logger"
	"webook/tag/repository"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"
)

func InitRepository(dao dao.TagDao, cache cache.TagCache, l logger.LoggerV1) repository.TagRepository {
	return repository.NewCachedTagRepository(dao, cache, l)
}
