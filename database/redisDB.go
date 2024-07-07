package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisDBStore struct {
	RedisDBClient *redis.Client
}

func NewRedisDBInstance() *RedisDBStore {

	client := redis.NewClient(&redis.Options{Addr: "redisDB:6379", Password: "", DB: 0})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal("error while connecting to redis database")
	} else {
		log.Println("redis client initialized.")
	}

	return &RedisDBStore{RedisDBClient: client}
}

func (rdb *RedisDBStore) UserConnected(userId string, endServerAddress string) error {

	err := rdb.RedisDBClient.Set(context.Background(), userId, endServerAddress, 0).Err()

	return err
}

func (rdb *RedisDBStore) UserDisconnected(userId string) error {

	_, err := rdb.RedisDBClient.Del(context.Background(), userId).Result()

	if err != nil {
		return err
	}

	log.Println(userId + " record has been deleted from redis database.")

	return nil
}

func (rdb *RedisDBStore) FindUserEndServerAddress(userId string) (string, error) {

	endServerAddress, err := rdb.RedisDBClient.Get(context.Background(), userId).Result()

	if err != nil {
		return "", err
	}

	return endServerAddress, nil

}
