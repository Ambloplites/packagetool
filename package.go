package packagetool

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"path"

	"github.com/Ambloplites/fsdump"
)

type backend interface {
	name() string
	probe(head [4]byte) bool
	read(reader *readerAtOffset) iter.Seq2[*FileMetadata, error]
	makeReader(metadata *FileMetadata, reader io.Reader) io.Reader
}

var backends = []backend{
	qarBackend{},
	barBackend{},
	marBackend{},
}

type PackageReaderSettings struct {
	DecryptMAR bool
}

type PackageReader struct {
	backend backend
	rd      io.ReaderAt
	offset  int64
}

func NewPackageReader(rd io.ReaderAt, offset int64, settings PackageReaderSettings) (*PackageReader, error) {
	var (
		backend backend
		head    [4]byte
	)

	if _, err := rd.ReadAt(head[:], offset); err != nil {
		return nil, err
	}
	found := false
	for _, backend = range backends {
		if backend.probe(head) {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("unable to detect format")
	}

	if b, ok := backend.(marBackend); ok && settings.DecryptMAR {
		b.useDecryption = true
		backend = b
	}

	return &PackageReader{
		rd:      rd,
		offset:  offset,
		backend: backend,
	}, nil
}

func (pr *PackageReader) Files() iter.Seq2[*FileMetadata, error] {
	return pr.backend.read(&readerAtOffset{
		rd:     pr.rd,
		offset: pr.offset + 4,
	})
}

func (pr *PackageReader) MakeFileReader(meta *FileMetadata) io.Reader {
	return pr.backend.makeReader(meta, io.NewSectionReader(pr.rd, meta.offset, meta.size))
}

func (pr *PackageReader) FileFormat() string {
	return pr.backend.name()
}

type Package struct {
	rd      *PackageReader
	files   []*FileMetadata
	fileMap map[string]*FileMetadata
}

var _ fs.FS = (*Package)(nil)

func NewPackage(rd *PackageReader) (_ *Package, err error) {
	p := &Package{
		rd:      rd,
		fileMap: map[string]*FileMetadata{},
	}

	for file, err := range rd.Files() {
		if err != nil {
			return nil, err
		}

		if file.size < 0 {
			return nil, fmt.Errorf("invalid file size: %d", file.size)
		}
		if file.offset < 0 {
			return nil, fmt.Errorf("invalid file offset: %d", file.offset)
		}
		if _, ok := p.fileMap[file.cleanPath]; ok {
			return nil, errors.New("duplicate filename: " + file.cleanPath)
		}

		p.fileMap[file.cleanPath] = file
		p.files = append(p.files, file)
	}

	return p, nil
}

func (p *Package) FileCount() int {
	return len(p.fileMap)
}

func (p *Package) Files() iter.Seq[*FileMetadata] {
	return func(yield func(*FileMetadata) bool) {
		for _, file := range p.files {
			if !yield(file) {
				return
			}
		}
	}
}

func (p *Package) Dump() chan fsdump.File {
	ch := make(chan fsdump.File, 1)
	closer := io.NopCloser(nil)
	go func() {
		for _, entry := range p.files {
			ch <- fsdump.File{
				Reader: p.rd.MakeFileReader(entry),
				Closer: closer,
				Path:   entry.cleanPath,
			}
		}
		close(ch)
	}()
	return ch
}

func (p *Package) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	entry, ok := p.fileMap[name]
	if !ok {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrNotExist,
		}
	}

	return &File{
		rd:   p.rd.MakeFileReader(entry),
		meta: entry,
	}, nil
}

func (p *Package) FindContents() string {
	for _, entry := range p.files {
		contents, name := path.Split(entry.cleanPath)
		if name == "allfiles.lst" {
			if contents == "" {
				contents = "."
			}
			return contents
		}
	}
	return ""
}

func (p *Package) ContentsFS() (fs.FS, error) {
	contents := p.FindContents()
	if contents == "" {
		return nil, errors.New("unable to find contents directory")
	}

	if i := len(contents) - 1; contents[i] == '/' {
		contents = contents[:i]
	}
	return fs.Sub(p, contents)
}
