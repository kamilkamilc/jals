package store

import (
	"context"

	"github.com/kamilkamilc/jals/config"
	"github.com/kamilkamilc/jals/model"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	rdb *redis.Client
	ctx context.Context
}

func InitializeRedisStorage(conf *config.Config) *RedisStorage {
	redisStorage := &RedisStorage{}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.RedisAddress,
		Password: "",
		DB:       conf.RedisDB,
	})
	redisStorage.rdb = redisClient
	redisStorage.ctx = context.Background()
	return redisStorage
}

func (rs *RedisStorage) SaveLink(link *model.Link) error {
	key := "links:" + link.ShortLink
	err := rs.rdb.HSet(rs.ctx, key, link.LinkInfo).Err()
	return err
}

func (rs *RedisStorage) RetrieveOriginalLink(shortLink string) (string, error) {
	key := "links:" + shortLink
	originalLink, err := rs.rdb.HGet(rs.ctx, key, "originalLink").Result()
	return originalLink, err
}

func (rs *RedisStorage) RetrieveLinkInfo(shortLink string) (*model.LinkInfo, error) {
	key := "links:" + shortLink
	linkInfo := &model.LinkInfo{}
	err := rs.rdb.HGetAll(rs.ctx, key).Scan(linkInfo)
	return linkInfo, err
}

func (rs *RedisStorage) IncrementClicks(shortLink string) error {
	key := "links:" + shortLink
	err := rs.rdb.HIncrBy(rs.ctx, key, "clicks", 1).Err()
	return err
}
