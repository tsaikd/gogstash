package filesystem

import (
	"os"
)

// FileInfo interface used so we can mock during tests
type FileInfo any

// FileSystem interface used so we can mock during tests
type FileSystem interface {
	Open(name string) (*os.File, error)
	Stat(name string) (FileInfo, error)
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}

// File interface used so we can mock during tests
type File interface {
	// io.Closer
	// io.Reader
	// io.ReaderAt
	// io.Seeker
	Sync() error
	Stat() (os.FileInfo, error)
	Write(b []byte) (n int, err error)
}
