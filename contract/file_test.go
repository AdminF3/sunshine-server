package contract

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileInTempDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "temp_dir_for_file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	osfile, err := ioutil.TempFile(dir, "temp")
	if err != nil {
		t.Fatal(err)
	}

	if n, err := osfile.Write([]byte("hello")); err != nil || n != 5 {
		t.Errorf("Wrote %d bytes with error: %q", n, err)
	}

	osfile.Close()
	defer os.Remove(osfile.Name())

	f, err := OpenFileInTempDir(osfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	var b = make([]byte, 8)
	if n, err := f.Read(b); err != nil || n != 5 {
		t.Errorf("Read %d bytes (%q) with error: %q", n, b, err)
	}

	// confirm we're not shadowing *os.File and implement the interfaces
	// one would expect for it to implement.
	var (
		_ io.Reader = f
		_ io.Writer = f
		_ io.Seeker = f
		_ io.Closer = f
	)

	// this should close the file and delete its directory
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err = OpenFileInTempDir(osfile.Name()); !os.IsNotExist(err) {
		t.Fatal("OpenFileInTempDir should not exist after Close")
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		fmt.Printf("Temporary directory should've been removed after Close")
	}
}
