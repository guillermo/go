package messagestore

import (
	"fmt"
	"testing"
)

type Expectations struct {
	Name        string
	T           *testing.T
	Messages    []string
	First       int
	Last        int
	Size        int
	nextPointer int
}

func (e Expectations) Verify(ms *MessageStore) {
	t := e.T
	msgs := ms.Messages()
	if len(msgs) != len(e.Messages) {
		t.Error(e.Name, "Expected", e.Messages, "Get", msgs)
	}
	for i, expected := range e.Messages {
		if msgs[i].(string) != expected {
			t.Error(e.Name, "Expected", e.Messages, "Get", msgs)
		}
	}

	if ms.First() != e.First {
		t.Error(e.Name, "Expected First() to be", e.First, "Get", ms.First())
	}
	if ms.Last() != e.Last {
		t.Error(e.Name, "Expected Last() to be", e.Last, "Get", ms.Last())
	}
	if ms.Size() != e.Size {
		t.Error(e.Name, "Expected Size() to be", e.Size, "Get", ms.Size())
	}
	if ms.nextPointer != e.nextPointer {
		t.Error(e.Name, "Expected nextPointer to be", e.nextPointer, "Get", ms.nextPointer)
	}
}

func TestMessageStore(t *testing.T) {

	ms := NewMessageStore(1)
	Expectations{"New", t, []string{}, 0, 0, 0, 0}.Verify(ms)
	ms.Push("hola")
	Expectations{"Push1", t, []string{"hola"}, 0, 0, 1, 0}.Verify(ms)
	ms.Push("adios")
	Expectations{"Push2", t, []string{"adios"}, 1, 1, 1, 0}.Verify(ms)
	ms.Push("hey")
	Expectations{"Push3", t, []string{"hey"}, 2, 2, 1, 0}.Verify(ms)

	ms = NewMessageStore(2)
	Expectations{"New", t, []string{}, 0, 0, 0, 0}.Verify(ms)
	ms.Push("hola")
	Expectations{"Push1", t, []string{"hola"}, 0, 0, 1, 1}.Verify(ms)
	ms.Push("adios")
	Expectations{"Push2", t, []string{"hola", "adios"}, 0, 1, 2, 0}.Verify(ms)
	ms.Push("hey")
	Expectations{"Push3", t, []string{"adios", "hey"}, 1, 2, 2, 1}.Verify(ms)

	ms = NewMessageStore(3)
	Expectations{"New", t, []string{}, 0, 0, 0, 0}.Verify(ms)
	ms.Push("hola")
	Expectations{"Push1", t, []string{"hola"}, 0, 0, 1, 1}.Verify(ms)
	ms.Push("adios")
	Expectations{"Push2", t, []string{"hola", "adios"}, 0, 1, 2, 2}.Verify(ms)
	ms.Push("hey")
	Expectations{"Push3", t, []string{"hola", "adios", "hey"}, 0, 2, 3, 0}.Verify(ms)
	ms.Push("T")
	Expectations{"Push3", t, []string{"adios", "hey", "T"}, 1, 3, 3, 1}.Verify(ms)
}

type GetExpectations struct {
	Name         string
	T            *testing.T
	Expectations []interface{}
}

func (e GetExpectations) Verify(ms *MessageStore) {
	t := e.T
	for i, expectation := range e.Expectations {

		result, err := ms.Get(i - 1)

		switch val := expectation.(type) {
		case nil: // We expect Get(i) to return nil and IndexOutOfRange
			if result != nil || err != IndexOutOfRange {
				t.Error(e.Name, "For ", ms.Messages(), "First", ms.First(),
					"We expect that Get(", i-1, ") to return nil (", result, ") and error",
					err)
			}
		case string:
			if val != expectation || err != nil {
				t.Error(e.Name, "For ", ms.Messages(), "First", ms.First(),
					"We expect that Get(", i-1, ") to return", expectation, "get", val, "with err",
					err)
			}
		}
	}
}

func TestGet(t *testing.T) {
	ms := NewMessageStore(2)
	GetExpectations{"Empty", t, []interface{}{nil, nil, nil, nil, nil, nil}}.Verify(ms)
	ms.Push("hola")
	GetExpectations{"hola", t, []interface{}{nil, "hola", nil, nil, nil, nil}}.Verify(ms)
	ms.Push("mundo")
	GetExpectations{"hola", t, []interface{}{nil, "hola", "mundo", nil, nil, nil}}.Verify(ms)
	ms.Push("cruel")
	GetExpectations{"hola", t, []interface{}{nil, nil, "mundo", "cruel", nil, nil, nil}}.Verify(ms)
}

func ExampleMessageStore() {
	ms := NewMessageStore(2)
	fmt.Println("New       ", ms.Messages())

	ms.Push("hello")
	fmt.Println("Push hello", ms.Messages())

	ms.Push("world")
	fmt.Println("Push world", ms.Messages())

	ms.Push("!")
	fmt.Println("Push !    ", ms.Messages())

	fmt.Println("First", ms.First(), "Last", ms.Last(), "Size", ms.Size())
	// Output:
	// New        []
	// Push hello [hello]
	// Push world [hello world]
	// Push !     [world !]
	// First 1 Last 2 Size 2

}

// to String Slice
func toSS(s []interface{}) (data []string) {
	data = make([]string, len(s))
	for i, v := range s {
		data[i] = v.(string)
	}
	return data
}

func compare(t *testing.T, test string, a []interface{}, b []string) {
	if len(a) != len(b) {
		t.Fatal(test, "Get len", toSS(a), "Expect len", b)
		for i, v := range toSS(a) {
			if v != b[i] {
				t.Error(test, "Get", v, "Expect", b)
			}
		}
	}
}

func TestRange(t *testing.T) {
	ms := NewMessageStoreWithFirst(2, 10)

	compare(t, "t1", ms.Range(0, 0), []string{})
	compare(t, "t2", ms.Range(10, 10), []string{})

	// One item
	ms.Push("A") // Should have index 10

	compare(t, "t3", ms.Range(0, 0), []string{})
	compare(t, "t4", ms.Range(10, 10), []string{})
	compare(t, "t5", ms.Range(10, 11), []string{"A"})

	ms.Push("B") // Should have index 10

	compare(t, "t6", ms.Range(0, 0), []string{})
	compare(t, "t7", ms.Range(10, 10), []string{})
	compare(t, "t8", ms.Range(10, 11), []string{"A"})
	compare(t, "t9", ms.Range(10, 12), []string{"A", "B"})
	compare(t, "ta", ms.Range(11, 12), []string{"B"})
	compare(t, "tb", ms.Range(11, 16), []string{"B"})
	compare(t, "tc", ms.Range(12, 16), []string{})

}

func TestFrom(t *testing.T) {
	ms := NewMessageStoreWithFirst(2, 10)

	compare(t, "t1", ms.From(0), []string{})
	compare(t, "t2", ms.From(10), []string{})

	// One item
	ms.Push("A") // Should have index 10

	compare(t, "t3", ms.From(0), []string{})
	compare(t, "t4", ms.From(10), []string{"A"})
	compare(t, "t5", ms.From(10), []string{"A"})

	ms.Push("B") // Should have index 10

	compare(t, "t6", ms.From(0), []string{})
	compare(t, "t7", ms.From(10), []string{"A", "B"})
	compare(t, "t8", ms.From(10), []string{"A", "B"})
	compare(t, "t9", ms.From(10), []string{"A", "B"})
	compare(t, "ta", ms.From(11), []string{"B"})
	compare(t, "tb", ms.From(11), []string{"B"})
	compare(t, "tc", ms.From(12), []string{})

}

func ExampleMessageStoreGet() {
	ms := NewMessageStore(1)
	val, err := ms.Get(0)
	if err != nil {
		fmt.Println("Get(0)", val, err)
	}

	ms.Push("hello")
	ms.Push("world")

	val, err = ms.Get(1)
	if err == nil {
		fmt.Println("Get(1)", val, err)
	}

	// Output:
	// Get(0) <nil> Index Out of Range
	// Get(1) world <nil>
}
