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
	lock       *utils.StringKeyLock
}

func CreateGenericCache[T interface{}](cacheId string, fetchItem func(id string) T, expiration time.Duration) GenericCache[T] {
	return GenericCache[T]{
		cacheId:    cacheId,
		fetchItem:  fetchItem,
		expiration: expiration,
		lock:       utils.NewStringKeyLock(),
	}
}

func (c *GenericCache[T]) getCacheKey(id string) string {
	return c.cacheId + ":" + id
}

func (c *GenericCache[T]) Delete(id string) error {
	c.lock.Lock(id)
	defer c.lock.Unlock(id)

	var err = utils.GetRedisClient().Del(context.Background(), c.getCacheKey(id)).Err()
	return err
}

func (c *GenericCache[T]) Set(id string, val T) error {
	c.lock.Lock(id)
	defer c.lock.Unlock(id)

	return c._set(id, val)
}

func (c *GenericCache[T]) _set(id string, val T) error {
	// Parse T to json
	var jsonVal, err = json.Marshal(val)
	if err != nil {
		return err
	}

	err = utils.GetRedisClient().Set(context.Background(), c.getCacheKey(id), jsonVal, c.expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *GenericCache[T]) Get(id string) (*T, error) {
	c.lock.Lock(id)

	var jsonVal, err = utils.GetRedisClient().Get(context.Background(), c.getCacheKey(id)).Result()

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

	// Parse json to T
	var val T
	err = json.Unmarshal([]byte(jsonVal), &val)

	if err != nil {
		log.Printf("%+v\n", err)
		return nil, err
	}

	c.lock.Unlock(id)
	return &val, nil
}
