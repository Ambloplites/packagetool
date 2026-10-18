package packagetool

import (
	"io"
	"io/fs"
	"path"
	"time"
)

type FileMetadata struct {
	realPath  string
	cleanPath string
	size      int64
	offset    int64
}

var _ fs.FileInfo = (*FileMetadata)(nil)

func (m *FileMetadata) RealPath() string {
	return m.realPath
}

func (m *FileMetadata) CleanPath() string {
	return m.cleanPath
}

func (m *FileMetadata) Name() string {
	_, name := path.Split(m.cleanPath)
	return name
}

func (m *FileMetadata) Size() int64 {
	return m.size
}

func (m *FileMetadata) Mode() fs.FileMode {
	return 0555
}

func (m *FileMetadata) IsDir() bool {
	return false
}

func (m *FileMetadata) ModTime() time.Time {
	return time.Time{}
}

func (m *FileMetadata) Type() fs.FileMode {
	return 0
}

func (m *FileMetadata) Info() (fs.FileInfo, error) {
	return m, nil
}

func (m *FileMetadata) Sys() any {
	return nil
}

type File struct {
	rd   io.Reader
	meta *FileMetadata
}

var _ fs.File = (*File)(nil)

func (f *File) Read(b []byte) (int, error) {
	return f.rd.Read(b)
}

func (f *File) Close() error {
	return nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f.meta, nil
}
