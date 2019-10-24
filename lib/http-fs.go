package lib

import (
	"bytes"
	"errors"
	"os"
	"path"
	"strings"
	"time"
)

type dir struct {
	group *FileGroup
	prefix string
}

type file struct {
	name     string
	dir    *dir
	buffer   *bytes.Reader
	children []string
	exists bool
}

func list(d *dir) (bool, []string) {
	exists := false
	all := d.group.Entries()
	children := []string{}
	for _, e := range all {
		if e == d.prefix {
			exists = true
		}
		if strings.HasPrefix(e, d.prefix + "/") {
			children = append(children, e)
		}
	}
	if !exists && len(children) == 0 {
		return false, nil
	}
	return exists, children
}

func (f *FileGroup) Open(name string) (*file, error) {
	exists, children := list(&dir{f, name})
	if children == nil {
		return nil, os.ErrNotExist
	}
	return &file{name, &dir{f, path.Dir(name)}, nil, children, exists}, nil
}

func (f *file) open() error {
	if !f.exists {
		return errors.New("can't read from a directory")
	}
	if f.buffer == nil {
		b := f.dir.group.MustBytes(f.name)
		f.buffer = bytes.NewReader(b)
	}
	return nil
}

func (f *file) Read(p []byte) (int, error) {
	if err := f.open(); err != nil {
		return 0, err
	}
	return f.buffer.Read(p)
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	if err := f.open(); err != nil {
		return 0, err
	}
	return f.buffer.Seek(offset, whence)
}

func (f *file) Close() error {
	f.buffer = nil
	return nil
}

func (f *file) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	if !f.IsDir() {
		return nil, errors.New("not a directory")
	}
	r := make([]os.FileInfo, len(f.children))
	for i, c := range f.children {
		exists, children := list(&dir{f.dir.group, c})
		r[i] = &file{c, &dir{f.dir.group, f.name}, nil, children, exists}
	}
	return r, nil
}

// Name returns the base name of the file
func (f *file) Name() string {
	return path.Base(f.name)
}

// Size returns the length in bytes for regular files; 0 for directories, -1 for inaccessible/nonexistant files
func (f *file) Size() int64 {
	if f.IsDir() {
		return 0
	}
	if err := f.open(); err != nil {
		return -1
	}
	return f.buffer.Size()
}

// Mode returns the file mode bits, which is always 0444 for Mewn
func (f *file) Mode() os.FileMode {
	var m os.FileMode
	if f.IsDir() {
		m |= os.ModeDir
	}
	return m
}

// ModTime returns the modification time, which is always the current time for Mewn
func (f *file) ModTime() time.Time {
	return time.Now()
}

// IsDir returns if the file is a directory
func (f *file) IsDir() bool {
	return f.children != nil && len(f.children) > 0
}

// Sys returns the underlying data source (is always nil for Mewn)
func (f *file) Sys() interface{} {
	return nil
}
