package main

import (
	memcached "./go-memcached"
)

type Cache struct{}

func (c *Cache) Get(key string) (item *memcached.Item, err error) {
	if key == "hello" {
		item = &memcached.Item{
			Key:   key,
			Value: []byte("world"),
		}
		return item, nil
	}
	return nil, memcached.NotFound
}

func main() {
	server := memcached.NewServer(":11211", &Cache{})
	server.ListenAndServe()
}
