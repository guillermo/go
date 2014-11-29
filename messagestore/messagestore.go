// Package messagesstore implements an indexed, push only, circular buffer.
//
// New messages will be added until it reach the maxium size of the store.
// After the maxium size is reach old messages are remove.
package messagestore

import (
	"errors"
	"sync"
)

type MessageStore struct {
	mu          sync.RWMutex
	first       int // Index of the first element. It start in 0 unless set otherwise.
	msgs        []interface{}
	nextPointer int
}

// NewMessageStore creates a new MessageStore with a maxium size.
// If size is lower than 1 it will panic.
func NewMessageStore(size int) *MessageStore {
	if size <= 0 {
		panic("Initialize a buffer with")
	}
	ms := &MessageStore{
		msgs: make([]interface{}, 0, size),
	}
	return ms
}

// NewMessageStoreWithFirst is like NewMessageStore but allows specify the
// index of the first element.
func NewMessageStoreWithFirst(size, first int) *MessageStore {
	ms := NewMessageStore(size)
	ms.first = first
	return ms
}

// Push will add new messages to the store.
// If the store is full it will override old messages.
func (ms *MessageStore) Push(msg interface{}) {
	defer ms.inc()
	defer ms.mu.Unlock()
	if len(ms.msgs) != cap(ms.msgs) {
		ms.mu.Lock()
		ms.msgs = append(ms.msgs, msg)
		return
	}
	ms.mu.Lock()
	ms.msgs[ms.nextPointer] = msg
	ms.first += 1
}

func (ms *MessageStore) inc() {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if ms.nextPointer+1 >= cap(ms.msgs) {
		ms.nextPointer = 0
		return
	}
	ms.nextPointer += 1
}

// Last return the index of the last element or First in case there is no
// elements.
func (ms *MessageStore) Last() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if len(ms.msgs) == 0 {
		return ms.first
	} else {
		return ms.first + len(ms.msgs) - 1
	}
}

// First return the index of the first element.
func (ms *MessageStore) First() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.first
}

// Messages return all the messages in the order they were push.
func (ms *MessageStore) Messages() []interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if len(ms.msgs) != cap(ms.msgs) {
		return ms.msgs[:len(ms.msgs)]
	}
	return append(ms.msgs[ms.nextPointer:], ms.msgs[:ms.nextPointer]...)
}

// Size return the current size of the store.
func (ms *MessageStore) Size() int {
	return len(ms.msgs)
}

var (
	IndexOutOfRange = errors.New("Index Out of Range")
)

func (ms *MessageStore) Get(index int) (interface{}, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if index < ms.first || index >= ms.first+len(ms.msgs) || len(ms.msgs) == 0 {
		return nil, IndexOutOfRange
	}

	offset := index - ms.first
	if ms.nextPointer+offset <= cap(ms.msgs) {
		return ms.msgs[offset], nil
	}
	return ms.msgs[offset-cap(ms.msgs)], nil
}
