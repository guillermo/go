// Package messagesstore implements an indexed, push only, circular buffer.
//
// New messages will be added until it reach the maxium size of the store.
// After the maxium size is reach old messages are remove.
package messagestore

import (
	"errors"
	"sync"
)

// MessageStore is a buffered indexed with fixed maxium size FIFO.
// By default the first element is indexed as 0.
type MessageStore struct {
	mu          sync.RWMutex
	first       int // Index of the first element. It start in 0 unless set otherwise.
	msgs        []interface{}
	nextPointer int
}

var (
	IndexOutOfRange = errors.New("Index Out of Range")
)

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

// Get return the element with the specify index. If the index is out of range
// IndexOutOfRange is returned as an error.
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

// Range return an slice with the elements which index is included between the
// maxium and minimum. [from, to). The returned slice will have a maximum of
// to-from elements.
func (ms *MessageStore) Range(from int, to int) []interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if len(ms.msgs) == 0 || to < ms.first || from > (ms.first+len(ms.msgs)) {
		return []interface{}{}
	}

	first := from - ms.first
	last := to - ms.first
	if last > len(ms.msgs) {
		last = len(ms.msgs)
	}
	return ms.Messages()[first:last]
}

// From return an slice with the elements which index is bigger than from
func (ms *MessageStore) From(from int) []interface{} {
	return ms.Range(from, from+len(ms.msgs))
}
