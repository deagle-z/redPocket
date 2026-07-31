package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"time"
)

var RD *redis.Client

type OwnedRedisLock struct {
	key   string
	token string
	ttl   time.Duration
}

var releaseOwnedLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
end
return 0
`)

var renewOwnedLockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("pexpire", KEYS[1], ARGV[2])
end
return 0
`)

func InitRD() (err error) {
	RD = redis.NewClient(&redis.Options{
		Addr:     GlobalConfig.Redis.Host,
		Password: GlobalConfig.Redis.Pass,
		DB:       GlobalConfig.Redis.Db,
	})
	_, err = RD.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Connect Redis server error:", err)
		return
	}
	return nil
}

func AcquireLock(lockKey string, seconds time.Duration) (bool, error) {
	result, err := RD.SetNX(context.Background(), lockKey, "1", seconds).Result()
	if err != nil {
		log.Printf("key:%s Connect Redis server error:", lockKey, err)
		return false, err
	}
	return result, nil
}

func IsKeyExistAndGetValue(lockKey string) (bool, string, error) {
	val, err := RD.Get(context.Background(), lockKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, val, nil
}

func ReleaseLock(lockKey string) error {
	_ = RD.Del(context.Background(), lockKey)
	return nil
}

func AcquireOwnedLock(lockKey string, ttl time.Duration) (*OwnedRedisLock, bool, error) {
	if RD == nil {
		return nil, false, errors.New("redis_not_available")
	}
	if lockKey == "" || ttl <= 0 {
		return nil, false, errors.New("owned_lock_invalid")
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, false, err
	}
	lock := &OwnedRedisLock{
		key:   lockKey,
		token: hex.EncodeToString(tokenBytes),
		ttl:   ttl,
	}
	acquired, err := RD.SetNX(context.Background(), lock.key, lock.token, ttl).Result()
	if err != nil {
		return nil, false, err
	}
	if !acquired {
		return nil, false, nil
	}
	return lock, true, nil
}

func (lock *OwnedRedisLock) Renew() error {
	if lock == nil || RD == nil {
		return errors.New("owned_lock_not_available")
	}
	result, err := renewOwnedLockScript.Run(
		context.Background(),
		RD,
		[]string{lock.key},
		lock.token,
		strconv.FormatInt(lock.ttl.Milliseconds(), 10),
	).Int64()
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("owned_lock_lost")
	}
	return nil
}

func (lock *OwnedRedisLock) Release() error {
	if lock == nil || RD == nil {
		return nil
	}
	_, err := releaseOwnedLockScript.Run(context.Background(), RD, []string{lock.key}, lock.token).Result()
	return err
}

func GetRdInt64(key string, defaultValue int64) (result int64) {
	dataStr := ""
	data := RD.Get(context.Background(), key)
	if data != nil && data.Err() == nil {
		dataStr = data.Val()
	}
	if dataStr != "" {
		num, err := strconv.ParseInt(dataStr, 10, 64)
		if err == nil {
			return num
		}
	}
	return defaultValue
}

func GetRdString(key string, defaultValue string) (result string) {
	data := RD.Get(context.Background(), key)
	if data != nil && data.Err() == nil {
		defaultValue = data.Val()
	}
	return defaultValue
}

func RdTimerSet(timerKey string, timerTime int64, data string) (err error) {
	return RD.ZAdd(context.Background(), timerKey, &redis.Z{
		Score:  float64(timerTime),
		Member: data,
	}).Err()
}

func RdTimerGet(timerKey string) (result []string, err error) {
	now := time.Now().Unix()
	tasks, err := RD.ZRangeByScore(context.Background(), timerKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    fmt.Sprintf("%d", now),
		Offset: 0,
		Count:  5000,
	}).Result()
	if err != nil {
		return result, err
	}
	for _, taskID := range tasks {
		result = append(result, taskID)
	}
	return result, err
}

func RdTimerDel(timerKey string, taskId string) (err error) {
	_, err = RD.ZRem(context.Background(), timerKey, taskId).Result()
	return err
}
