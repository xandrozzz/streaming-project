package database

import (
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"log"
	"streaming-project/models"
)

type OrderCache struct {
	orders *cache.Cache
}

const (
	defaultExpiration = cache.NoExpiration
	purgeTime         = cache.NoExpiration
)

func NewCache() *OrderCache {
	Cache := cache.New(defaultExpiration, purgeTime)
	return &OrderCache{
		orders: Cache,
	}
}

func (c *OrderCache) Read(id string) (item *models.Order, ok bool) {
	product, ok := c.orders.Get(id)
	if ok {
		log.Println("from cache")
		res, err := json.Marshal(product.(models.Order))
		if err != nil {
			log.Fatal("Error")
		}
		var order models.Order
		err = json.Unmarshal(res, &order)
		if err != nil {
			return nil, false
		}
		return &order, true
	}
	return nil, false
}

func (c *OrderCache) Update(id string, order models.Order) {
	c.orders.Set(id, order, cache.DefaultExpiration)
}
