package messagesubscribe

import (
	"fmt"
	"sync"
	"testing"
)

var data string
var wg sync.WaitGroup

func TestMessageBroker(t *testing.T) {
	var wg, wg2, wg3 sync.WaitGroup

	b := NewMessageBrokerWithChannel(2, make(chan (interface{})))

	b.C <- "Hello"
	b.C <- "World"
	wg.Add(1)
	wg2.Add(1)
	wg3.Add(1)
	go func() {
		s := b.SubscribeFrom(1)
		msg, ok := <-s.C
		if !ok {
			t.Error("Channel should be open")
		}
		if msg.Index != 1 {
			t.Error("The message index should be 1. Got", msg.Index)
		}
		if msg.Data.(string) != "World" {
			t.Error("The message index should be World. Got", msg.Data.(string))
		}
		wg.Done()

		msg, ok = <-s.C
		if !ok {
			t.Error("Channel should be open")
		}
		if msg.Index != 2 {
			t.Error("The message index should be 2. Got", msg.Index)
		}
		if msg.Data.(string) != "Bye" {
			t.Error("The message index should be Bye. Got", msg.Data.(string))
		}
		wg2.Done()

		_, ok = <-s.C
		if ok {
			t.Error("The channel should be close", ok)
		}
		wg3.Done()

	}()
	wg.Wait()

	b.C <- "Bye"
	t.Log("Waiting for wg2")
	wg2.Wait()

	close(b.C)
	t.Log("Waiting for wg3")
	wg3.Wait()
	t.Log("Done")
}

func ExampleNewMessageBroker() {
	var wg, wg2 sync.WaitGroup

	b := NewMessageBroker(2)

	b.C <- "Hello"
	b.C <- "World"
	b.C <- "!"

	wg.Add(1)
	wg2.Add(1)
	go func() {
		wg2.Done()
		s := b.SubscribeFrom(2)
		for msg := range s.C {
			fmt.Println("Message Index:", msg.Index, "Content:", msg.Data.(string))
		}
		wg.Done()
	}()
	wg2.Wait()

	b.C <- ":-P"

	close(b.C)
	wg.Wait()

	// Output:
	// Message Index: 2 Content: !
	// Message Index: 3 Content: :-P
}
