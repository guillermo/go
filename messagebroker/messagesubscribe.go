// Package messagesubscribe is a publis subscription model based in
// messagestore where each message gets an index (autoincrement) and
// subscribers can decide which should be the first message to get from the
// history. Server Sent Events is a good use case for MessageBroker
package messagebroker

import (
	. "github.com/guillermo/go/messagestore"
)

// MessageBroker is the main broker. Messages can be published through
// MessageBroker.C
// It have to be initialized, automatically by NewMessageBroker, or manually by
// calling Start.
// It could be stoped by closing the publish channel C.
type MessageBroker struct {
	ms              *MessageStore
	C               chan (interface{}) // Publish channel
	size            int                // Total number of messages to store
	subscribeChan   chan (*Subscription)
	unsubscribeChan chan (*Subscription)
	subscriptions   []*Subscription
}

// Subscription will receive all old stored messages with an index bigger than
// the one provide during the creation of the subscription, and also new
// messages. The channel C will have a default buffer of the same size of the
// current total number of messages to prevent blocking the broker. The channel
// C will be close on unsubscribe or on broker stop (The publish channel gets
// close).
type Subscription struct {
	C      chan (Message)
	first  int
	broker *MessageBroker
}

// Message is received structure in the subscription channel.
type Message struct {
	Index int
	Data  interface{}
}

// NewMessageBroker creates a new Broker with capacity for up to
// _size_ messages.
//
// The publish channel have a message buffer of 1024 by default.
//
// To stop the broker is enought to close the channel C. Ensure that all the
// subscriber waits until the subscription channel is also close.
func NewMessageBroker(size int) *MessageBroker {
	c := make(chan (interface{}), 1024)
	return NewMessageBrokerWithChannel(size, c)
}

// NewMessageBrokerWithChannel creates a new Broker with the specify
// channel. See NewMessageBroker.
func NewMessageBrokerWithChannel(size int, channel chan (interface{})) *MessageBroker {
	b := &MessageBroker{
		size:            size,
		C:               channel,
		ms:              NewMessageStore(size),
		subscribeChan:   make(chan (*Subscription)),
		unsubscribeChan: make(chan (*Subscription)),
		subscriptions:   make([]*Subscription, 0),
	}
	go b.loop()
	return b
}

// SubscribeFrom creates a new subscription that will receive all the previous
// messages with an index bigger than _first_ and all the new messages until
// the publish channel is close or the subscription is cancel through
// Unsubscribe(). Once that happends the channel C is close.
func (b *MessageBroker) SubscribeFrom(first int) *Subscription {
	//msgs := b.ms.From(s.first)
	//s.C = make(chan (Message), len(msgs))
	s := &Subscription{
		first:  first,
		C:      make(chan (Message)),
		broker: b,
	}

	go func() {
		b.subscribeChan <- s
	}()
	return s
}

// Unsubscribe will cancel the subscriptions. Messages should still arrive and
// you must to wait until the broker closes the channel.
func (s *Subscription) Unsubscribe() {
	s.broker.unsubscribeChan <- s
}

func (b *MessageBroker) loop() {
MainLoop:
	for {
		select {
		case msg, open := <-b.C:
			if !open {
				for _, s := range b.subscriptions {
					close(s.C)
				}
				b.subscriptions = nil
				break MainLoop
			}
			b.ms.Push(msg)
			last := b.ms.Last()
			for _, s := range b.subscriptions {
				s.C <- Message{last, msg}
			}
		case s := <-b.subscribeChan:
			b.subscriptions = append(b.subscriptions, s)
			msgs := b.ms.From(s.first)
			for i, msg := range msgs {
				s.C <- Message{s.first + i, msg}
			}
		case subscription := <-b.unsubscribeChan:

			for i, s := range b.subscriptions {
				if s == subscription {
					b.subscriptions = append(b.subscriptions[:i], b.subscriptions[i+1:]...)
					close(s.C)
				}
			}
		}
	}
}
