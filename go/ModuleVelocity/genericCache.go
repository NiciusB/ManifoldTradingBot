package modulevelocity

import (
	"ManifoldTradingBot/utils"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type GenericCache[T interface{}] struct {
	cacheId    string
	fetchItem  func(id string) T
	expiration time.Duration
	renewAfter time.Duration
	lock       *utils.StringKeyLock
}

func CreateGenericCache[T interface{}](cacheId string, fetchItem func(id string) T, expiration time.Duration, renewAfter time.Duration) GenericCache[T] {
	return GenericCache[T]{
		cacheId:    cacheId,
		fetchItem:  fetchItem,
		expiration: expiration,
		renewAfter: renewAfter,
		lock:       utils.NewStringKeyLock(),
	}
}

func (c *GenericCache[T]) getCacheKey(id string) string {
	return c.cacheId + ":" + id
}

func (c *GenericCache[T]) Delete(id string) error {
	c.lock.Lock(id)
	defer c.lock.Unlock(id)

	var err = utils.RedisClient.Del(context.Background(), c.getCacheKey(id)).Err()
	return err
}

func (c *GenericCache[T]) Set(id string, val T) error {
	c.lock.Lock(id)
	defer c.lock.Unlock(id)

	return c._set(id, val)
}

func (c *GenericCache[T]) Renew(id string) error {
	err := c.Delete(id)
	if err != nil {
		return err
	}

	_, err = c.Get(id)
	if err != nil {
		return err
	}

	return nil
}

func (c *GenericCache[T]) _set(id string, val T) error {
	// Parse T to json
	var jsonVal, err = json.Marshal(val)
	if err != nil {
		return err
	}

	err = utils.RedisClient.Set(context.Background(), c.getCacheKey(id), jsonVal, c.expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *GenericCache[T]) Get(id string) (*T, error) {
	c.lock.Lock(id)

	// Get data from redis
	var jsonVal, err = utils.RedisClient.Get(context.Background(), c.getCacheKey(id)).Result()

	// If redis did not contain the data, fetch new and save it
	if err == redis.Nil {
		var freshItem = c.fetchItem(id)

		go func() {
			c._set(id, freshItem)
			c.lock.Unlock(id)
		}()

		return &freshItem, nil
	}

	if err != nil {
		c.lock.Unlock(id)
		return nil, err
	}

	// Renew if TTL is small enough, we will return old data for this request but it'll be renewed for next
	go func() {
		ttl, err := utils.RedisClient.TTL(context.Background(), c.getCacheKey(id)).Result()
		if err != nil {
			return
		}

		if ttl <= c.renewAfter {
			c.Renew(id)
		}
	}()

	// Parse json to T
	var val T
	err = json.Unmarshal([]byte(jsonVal), &val)

	if err != nil {
		log.Printf("%+v\n", err)
		c.lock.Unlock(id)
		return nil, err
	}

	c.lock.Unlock(id)
	return &val, nil
}
