package packagetool

import (
	"errors"
	"io"
	"strings"
)

type readerAtOffset struct {
	rd     io.ReaderAt
	offset int64
}

func (rao *readerAtOffset) Read(b []byte) (n int, err error) {
	n, err = rao.rd.ReadAt(b, rao.offset)
	rao.offset += int64(n)
	return
}

const (
	removeDotSlash   = (1 << 2)
	replaceBackslash = (1 << 3)
)

func normalizePath(rawPath []byte, options int) (cleanPath, realPath string, _ error) {
	if len(rawPath) == 0 {
		return "", "", errors.New("path is empty")
	}

	for i := range rawPath {
		if rawPath[i] == 0 {
			realPath = string(rawPath[:i])
			break
		}
	}
	if len(realPath) == 0 {
		return "", "", errors.New("path is not null-terminated")
	}
	cleanPath = realPath

	if options&replaceBackslash != 0 {
		cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
	}

	if cleanPath[0] == '/' {
		cleanPath = cleanPath[1:]
	}

	if options&removeDotSlash != 0 {
		cleanPath, _ = strings.CutPrefix(cleanPath, "./")
	}

	if len(cleanPath) == 0 {
		return "", "", errors.New("path is empty after normalization")
	}

	return
}
