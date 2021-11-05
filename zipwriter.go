package shp

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// byteBuf provides a writeable + seekable byte buffer
type byteBuf struct {
	buf []byte
	pos int
}

func (b *byteBuf) Len() int { return len(b.buf) }

// GetReader returns an io.Reader for byteBuf contents
func (b *byteBuf) GetReader() io.Reader {
	return bytes.NewReader(b.buf)
}

// byteBuf implements io.Writer
func (b *byteBuf) Write(p []byte) (int, error) {
	extra := len(p) - len(b.buf[b.pos:])
	if extra > 0 {
		b.buf = append(b.buf, make([]byte, extra)...)
	}
	n := copy(b.buf[b.pos:], p)
	b.pos += n
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
	*Writer
	closed bool
}

// NewZipWriter returns an instantiated ZipWriter and the first error that was
// encountered. In case an error occurred the returned ZipWriter point will be nil.
// If filename does not end on ".shp" already, it will be treated as the basename
// for the file and the ".shp" extension will be appended to that name.
func NewZipWriter(filename string, t ShapeType) (*ZipWriter, error) {
	if strings.HasSuffix(strings.ToLower(filename), ".shp") {
		filename = filename[0 : len(filename)-4]
	}
	shp := &byteBuf{}
	shx := &byteBuf{}
	shp.Seek(100, io.SeekStart)
	shx.Seek(100, io.SeekStart)
	w := &Writer{
		filename:     filename,
		shp:          shp,
		shx:          shx,
		GeometryType: t,
	}
	return &ZipWriter{w, false}, nil
}

// SetFields sets field values in the DBF. This initializes the DBF buffer and
// should be used prior to writing any attributes.
func (w *ZipWriter) SetFields(fields []Field) error {
	if w.dbf != nil {
		return errors.New("Cannot set fields in existing dbf")
	}

	w.dbf = &byteBuf{}
	w.dbfFields = fields

	// calculate record length
	w.dbfRecordLength = int16(1)
	for _, field := range w.dbfFields {
		w.dbfRecordLength += int16(field.Size)
	}

	// header lengh
	w.dbfHeaderLength = int16(len(w.dbfFields)*32 + 33)

	// fill header space with empty bytes for now
	buf := make([]byte, w.dbfHeaderLength)
	binary.Write(w.dbf, binary.LittleEndian, buf)

	// write empty records
	for n := int32(0); n < w.num; n++ {
		w.writeEmptyRecord()
	}
	return nil
}

// Close closes the ZipWriter, finalizing file buffers with
// correct headers, and returning an error if already closed.
func (w *ZipWriter) Close() error {
	if w.closed {
		return fmt.Errorf("already closed")
	}

	// initialize dbf with empty fields if not already
	if w.dbf == nil {
		w.SetFields([]Field{})
	}

	w.closed = true
	w.Writer.Close()
	return nil
}

// Bytes builds and returns zip file bytes, closing the ZipWriter
// if not already closed.
func (w *ZipWriter) Bytes() ([]byte, error) {
	var (
		buf  = new(bytes.Buffer)
		zipw = zip.NewWriter(buf)
	)

	writeFile := func(path string, bb io.Reader) error {
		zf, err := zipw.Create(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(zf, bb)
		return err
	}

	w.Close()

	if err := writeFile(w.filename+".shp", w.shp.(*byteBuf).GetReader()); err != nil {
		return nil, err
	}

	if err := writeFile(w.filename+".shx", w.shx.(*byteBuf).GetReader()); err != nil {
		return nil, err
	}

	if w.dbf != nil {
		if err := writeFile(w.filename+".dbf", w.dbf.(*byteBuf).GetReader()); err != nil {
			return nil, err
		}
	}

	if err := zipw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
