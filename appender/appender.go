// Package Appender Serialize a bunch of bytes in messages that can be read independently
package appender

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
)

// DB just holds data common to the files
type DB struct {
}

// Open a specific file in the database
func (db *DB) Open(name string) (*File, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return &File{f: f}, nil
}

// Remove a file from the database
func (db *DB) Remove(name string) error {
	return os.Remove(name)
}

// File represent a basic file
type File struct {
	f *os.File
	m sync.Mutex
}

// Write at the end data into the file
func (f *File) Write(data []byte) (n int, err error) {
	f.m.Lock()
	defer f.m.Unlock()

	err = binary.Write(f.f, binary.LittleEndian, int64(len(data)))
	if err != nil {
		return 0, err
	}
	return f.f.Write(data)
}

// Close the file
func (f *File) Close() error {
	return f.f.Close()
}

// Iterator callback
type Iterator func(entry io.Reader)

// Blocks the file for reading all the content
func (f *File) Iterate(iterator Iterator) error {
	f.m.Lock()
	defer f.m.Unlock()
	// Use a section reader to increase performance

	_, err := f.f.Seek(0, 0)
	if err != nil {
		return err
	}
	defer f.f.Seek(0, 2)

	var size int64
	for {
		err := binary.Read(f.f, binary.LittleEndian, &size)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		iterator(&io.LimitedReader{R: f.f, N: size})
	}

	return err
}
