package appender

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func WriteAll(f *File, tests []string) error {

	for _, s := range tests {
		data := []byte(s)
		n, err := f.Write(data)
		if err != nil {
			return err
		}
		if n != len(data) {
			return fmt.Errorf("What the fuck!")
		}
	}

	return nil
}

func ReadAll(f *File) (tests []string, err error) {
	dataRead := []string{}
	err = f.Iterate(func(entry io.Reader) {
		readed, err := ioutil.ReadAll(entry)
		if err != nil {
			panic(err)
		}
		dataRead = append(dataRead, string(readed))
	})
	return dataRead, err
}

func Compare(a, b []string) error {
	if len(a) != len(b) {
		return fmt.Errorf("DIFFERENT:\nExpected: %s\nGet    : %s", b, a)
	}
	for i, v := range a {
		if v != b[i] {
			return fmt.Errorf("DIFFERENT:\nExpected: %s\nGet    : %s", b, a)
		}
	}
	return nil
}

func TestSomething2(t *testing.T) {
	db := &DB{}
	db.Remove("user321")
	defer db.Remove("user321")

	f, err := db.Open("user321")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	sample := []string{"hello", "world", "el", "mundo", "es", "un", "test test test"}
	err = WriteAll(f, sample)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	err = Compare(data, sample)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSomething(t *testing.T) {
	db := &DB{}
	db.Remove("user123")
	defer db.Remove("user123")

	f, err := db.Open("user123")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	d, err := ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(d) != 0 {
		t.Fatal("WTF")
	}

	testString := "hello world"
	n, err := f.Write([]byte(testString))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(testString) {
		t.Fatal("Expected to write", len(testString), "Wrote", n)
	}

	dataRead := []string{}
	err = f.Iterate(func(entry io.Reader) {
		readed, err := ioutil.ReadAll(entry)
		if err != nil {
			t.Error(err)
		}
		dataRead = append(dataRead, string(readed))
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(dataRead) != 1 {
		t.Fatal("Expected one entry", len(dataRead))
	}

	if dataRead[0] != "hello world" {
		t.Fatal("Not readed what was expected", dataRead[0])
	}

}
