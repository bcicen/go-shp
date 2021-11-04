package shp

import (
	"archive/zip"
	"fmt"
	"io"
)

// byteBuf provides a writeable + seekable byte buffer
type byteBuf struct {
	buf []byte
	pos int
}

func (b *byteBuf) Len() int { return len(b.buf) }

// byteBuf implements io.Writer
func (b *byteBuf) Write(p []byte) (int, error) {
	extra := len(p) - len(b.buf[b.pos:])
	if extra > 0 {
		b.buf = append(b.buf, make([]byte, extra)...)
	}
	n := copy(b.buf[b.pos:], p)
	return n, nil
}

// byteBuf implements io.Seeker
func (b *byteBuf) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		b.pos = int(offset)
	case io.SeekCurrent:
		b.pos += int(offset)
	case io.SeekEnd:
		b.pos = len(b.buf) + int(offset)
	default:
		return 0, fmt.Errorf("invalid whence")
	}

	if b.pos < 0 {
		return 0, fmt.Errorf("invalid seek %d", b.pos)
	}
	extra := b.pos - len(b.buf)
	if extra > 0 {
		b.buf = append(b.buf, make([]byte, extra)...)
	}
	return int64(b.pos), nil
}

// byteBuf implements io.Closer
func (b *byteBuf) Close() error {
	return nil
}

// ZipWriter provides an interface for writing shp and dbf files to a compressed ZIP archive.
type ZipWriter struct {
	*zip.Writer
}

func NewZipWriter() *ZipWriter {
	return nil
}
