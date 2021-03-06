package lru_cache

import (
	"errors"
	"sync"
	"time"
	"unsafe"
)

var ErrHugeItem = errors.New("item does not fit in cache")

type item struct {
	key        string
	value      []byte
	next       *item
	prev       *item
	expiration int64 // epoch nanosecond
}

type LRU struct {
	head  *item
	tail  *item
	table map[string]*item
	sync.Mutex
	size    int64 // bytes
	maxSize int64 // bytes
}

func GetLRLUCache(maxSize int64) *LRU {
	return &LRU{maxSize: maxSize}
}

func (lru *LRU) AddItem(key string, val []byte, expiration int64) error {
	lru.Lock()
	defer lru.Unlock()

	i := &item{key: key, value: val, next: nil, prev: nil, expiration: expiration}
	isize := int64(unsafe.Sizeof(i))
	if lru.size+isize < lru.maxSize {
		lru.RemoveExpiredItems()
	}
	for lru.size+isize < lru.maxSize {
		ok := lru.RemoveFromTail()
		if !ok {
			return ErrHugeItem
		}
	}
	// Fresh start
	if len(lru.table) == 0 {
		lru.head = i
		lru.tail = i
		lru.table = make(map[string]*item)
	} else {
		lru.tail.next = i
		temp := lru.tail
		lru.tail = i
		lru.tail.prev = temp
	}
	lru.table[key] = i
	lru.size += isize

	return nil
}

func (lru *LRU) GetItem(key string) (*item, bool) {
	i, ok := lru.table[key]
	if i.expiration > time.Now().UnixNano() {
		lru.MoveToHead(i.key)
		return i, ok
	} else {
		lru.RemoveItem(i.key)
		return nil, false
	}

}

func (lru *LRU) RemoveItem(key string) bool {
	lru.Lock()
	defer lru.Unlock()

	i, ok := lru.GetItem(key)
	if !ok {
		return false
	}
	if i.prev == nil {
		lru.head = i.next
		lru.head.prev = nil
	} else if i.next == nil {
		lru.tail = i.prev
		lru.tail.next = nil
	} else {
		i.prev.next = i.next
	}
	delete(lru.table, key)
	return true
}

func (lru *LRU) MoveToHead(key string) bool {
	lru.Lock()
	defer lru.Unlock()

	i, ok := lru.GetItem(key)
	if !ok {
		return false
	}
	if i.prev == nil {
		return true
	} else if i.next == nil {
		i.prev.next = nil
		i.next = lru.head
		i.prev = nil
		lru.head = i
	} else {
		i.prev.next = i.next
		i.next = lru.head
		i.prev = nil
		lru.head = i
	}
	return true
}

func (lru *LRU) RemoveFromTail() bool {
	return lru.RemoveItem(lru.tail.key)
}

func (lru *LRU) RemoveExpiredItems() {
	for k, v := range lru.table {
		if v.expiration < time.Now().UnixNano() {
			lru.RemoveItem(k)
		}
	}
	return
}
