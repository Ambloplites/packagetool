package packagetool

import (
	"encoding/binary"
	"io"
	"iter"
)

type qarBackend struct{}

func (qarBackend) name() string {
	return "QAR"
}

func (qarBackend) probe(head [4]byte) bool {
	return string(head[:]) == "QAR\x00"
}

func (qarBackend) read(rd *readerAtOffset) iter.Seq2[*FileMetadata, error] {
	return func(yield func(*FileMetadata, error) bool) {
		b := make([]byte, 144)

		// read file count
		if _, err := io.ReadFull(rd, b[:4]); err != nil {
			yield(nil, err)
			return
		}
		count := binary.LittleEndian.Uint32(b[:4])

		for range count {
			// read file metadata
			if _, err := io.ReadFull(rd, b); err != nil {
				yield(nil, err)
				return
			}
			size := int64(binary.LittleEndian.Uint64(b[136:]))
			cleanPath, realPath, err := normalizePath(b[:128], replaceBackslash|removeDotSlash)
			if err != nil {
				yield(nil, err)
				return
			}

			if !yield(&FileMetadata{
				realPath:  realPath,
				cleanPath: cleanPath,
				offset:    rd.offset,
				size:      size,
			}, nil) {
				return
			}
			rd.offset += size
		}
	}
}

func (qarBackend) makeReader(entry *FileMetadata, rd io.Reader) io.Reader {
	return rd
}
