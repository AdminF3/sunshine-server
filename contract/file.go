package contract

import (
	"os"
	"path/filepath"
)

// FileInTempDir represents an open file descriptor in (an assumed to be
// temporary) directory which will be removed when file is closed.
//
// Note that FileInTempDir is a drop-in replacement of *os.File.
type FileInTempDir struct {
	*os.File
}

// OpenFileInTempDir opens the named file for reading.
func OpenFileInTempDir(name string) (*FileInTempDir, error) {
	f, err := os.Open(name)
	return &FileInTempDir{
		File: f,
	}, err
}

// Close closes the file, rendering it unusable for I/O and deletes the
// directory where this file is located. It returns the result from
// underlaying's file.Close().
func (f FileInTempDir) Close() error {
	defer os.RemoveAll(filepath.Dir(f.File.Name()))
	return f.File.Close()
}
