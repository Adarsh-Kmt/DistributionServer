package database

import (
	"context"
	"log"
	"time"

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

func (rdb *RedisDBStore) UserConnected(username string, endServerAddress string) error {

	err := rdb.RedisDBClient.Get(context.Background(), username).Err()

	if err == redis.Nil {

		userRecordMap := map[string]any{"EndServerAddress": endServerAddress, "Online": true, "LastOnline": time.Now()}

		for k, v := range userRecordMap {

			err := rdb.RedisDBClient.HSet(context.Background(), username, k, v).Err()

			if err != nil {
				return err
			}
		}
	} else {
		err = rdb.RedisDBClient.HSet(context.Background(), username, "EndServerAddress", endServerAddress).Err()

	}

	return err
}

func (rdb *RedisDBStore) UserDisconnected(username string) error {

	// _, err := rdb.RedisDBClient.Del(context.Background(), userId).Result()

	// if err != nil {
	// 	return err
	// }

	// log.Println(userId + " record has been deleted from redis database.")

	userOfflineMap := map[string]any{"Online": false, "LastOnline": time.Now()}

	for k, v := range userOfflineMap {
		err := rdb.RedisDBClient.HSet(context.Background(), username, k, v).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (rdb *RedisDBStore) FindUserEndServerAddress(username string) (string, error) {

	endServerAddress, err := rdb.RedisDBClient.HGet(context.Background(), username, "EndServerAddress").Result()

	if err != nil {
		return "", err
	}

	return endServerAddress, nil

}
