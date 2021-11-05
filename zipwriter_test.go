package shp

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func zipUnzip(w *ZipWriter) error {
	path := fmt.Sprintf("%d.zip", time.Now().Nanosecond())
	//defer os.Remove(path)
	b, err := w.Bytes()
	if err != nil {
		return nil
	}

	if err := os.WriteFile(path, b, 0644); err != nil {
		return err
	}

	return exec.Command("unzip", path).Run()
}

func TestZipWritePoint(t *testing.T) {
	filename := filenamePrefix + "point"
	defer removeShapefile(filename)

	points := [][]float64{
		{0.0, 0.0},
		{5.0, 5.0},
		{10.0, 10.0},
	}

	shape, err := NewZipWriter(filename+".shp", POINT)
	t.Logf("%T %T %T", shape.shp, shape.shx, shape.dbf)
	require.NoError(t, err)
	for _, p := range points {
		shape.Write(&Point{p[0], p[1]})
		t.Logf("%T %T %T", shape.shp, shape.shx, shape.dbf)
	}
	require.NoError(t, zipUnzip(shape))

	shapes := getShapesFromFile(filename, t)
	if len(shapes) != len(points) {
		t.Error("Number of shapes read was wrong")
	}
	testPoint(t, points, shapes)
}

func TestZipWritePolyLine(t *testing.T) {
	filename := filenamePrefix + "polyline"
	defer removeShapefile(filename)

	points := [][]Point{
		{Point{0.0, 0.0}, Point{5.0, 5.0}},
		{Point{10.0, 10.0}, Point{15.0, 15.0}},
	}

	shape, err := NewZipWriter(filename+".shp", POLYLINE)
	require.NoError(t, err)

	l := NewPolyLine(points)

	lWant := &PolyLine{
		Box:       Box{MinX: 0, MinY: 0, MaxX: 15, MaxY: 15},
		NumParts:  2,
		NumPoints: 4,
		Parts:     []int32{0, 2},
		Points: []Point{{X: 0, Y: 0},
			{X: 5, Y: 5},
			{X: 10, Y: 10},
			{X: 15, Y: 15},
		},
	}
	assert.Equal(t, lWant, l)

	shape.Write(l)
	require.NoError(t, zipUnzip(shape))

	shapes := getShapesFromFile(filename, t)
	if len(shapes) != 1 {
		t.Error("Number of shapes read was wrong")
	}
	testPolyLine(t, pointsToFloats(flatten(points)), shapes)
}
