package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := c.items[key]; ok {
		cacheItem := item.Value.(*cacheItem)
		cacheItem.value = value
		c.queue.MoveToFront(item)
		return true
	}

	newCacheItem := &cacheItem{
		key:   key,
		value: value,
	}
	newListItem := c.queue.PushFront(newCacheItem)
	c.items[key] = newListItem

	if c.queue.Len() > c.capacity {
		back := c.queue.Back()
		if back != nil {
			backItem := back.Value.(*cacheItem)
			delete(c.items, backItem.key)
			c.queue.Remove(back)
		}
	}
	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := c.items[key]; ok {
		c.queue.MoveToFront(item)
		cacheItem := item.Value.(*cacheItem)
		return cacheItem.value, true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
