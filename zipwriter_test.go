package shp

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteBuf(t *testing.T) {
	b := &byteBuf{}

	n, err := b.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, b.Len())

	nn, err := b.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, 0, int(nn))
	assert.Equal(t, 5, b.Len())

	nn, err = b.Seek(128, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, 128, int(nn))
	assert.Equal(t, 128, b.Len())

	nn, err = b.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, 128, int(nn))
	assert.Equal(t, 128, b.Len())

	nn, err = b.Seek(128, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, 256, int(nn))
	assert.Equal(t, 256, b.Len())

	nn, err = b.Seek(128, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, 128, int(nn))
	assert.Equal(t, 256, b.Len())

	_, err = b.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 256, b.Len())

	nn, err = b.Seek(-8, io.SeekEnd)
	assert.NoError(t, err)
	assert.Equal(t, 248, int(nn))
	assert.Equal(t, 256, b.Len())

	_, err = b.Write([]byte("hellohello"))
	assert.NoError(t, err)
	assert.Equal(t, 258, b.Len())
}
