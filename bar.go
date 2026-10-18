package packagetool

import (
	"encoding/binary"
	"io"
	"iter"
)

type barBackend struct{}

func (barBackend) name() string {
	return "BAR"
}

func (barBackend) probe(head [4]byte) bool {
	return head[3] == 'A'
}

func (barBackend) read(rd *readerAtOffset) iter.Seq2[*FileMetadata, error] {
	return func(yield func(*FileMetadata, error) bool) {
		const filenameSize = 252
		b := make([]byte, 264)

		// read header
		if _, err := io.ReadFull(rd, b[:8]); err != nil {
			yield(nil, err)
			return
		}

		count := binary.LittleEndian.Uint16(b[6:])
		for range count {
			// read filename and the weird thing
			if _, err := io.ReadFull(rd, b[:filenameSize+4]); err != nil {
				yield(nil, err)
				return
			}
			cleanPath, realPath, err := normalizePath(b[:filenameSize], replaceBackslash)
			if err != nil {
				yield(nil, err)
				return
			}

			// for weird files
			if binary.LittleEndian.Uint32(b[filenameSize:]) != 3 {
				// skip
				rd.offset += 4
			}

			// read file size
			if _, err := io.ReadFull(rd, b[filenameSize:]); err != nil {
				yield(nil, err)
				return
			}
			size := int64(binary.LittleEndian.Uint64(b[filenameSize+4:]))

			if !yield(&FileMetadata{
				cleanPath: cleanPath,
				realPath:  realPath,
				offset:    rd.offset,
				size:      size,
			}, nil) {
				return
			}
			rd.offset += size
		}
	}
}

func (barBackend) makeReader(entry *FileMetadata, rd io.Reader) io.Reader {
	return rd
}
