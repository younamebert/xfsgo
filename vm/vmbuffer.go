package vm

import (
	"encoding/binary"
	"io"
)

type Buffer interface {
	ReadUint8() (CTypeUint8, error)
	ReadUint16() (CTypeUint16, error)
	ReadUint32() (CTypeUint32, error)
	ReadString(size int) (CTypeString, error)
	Write(p []byte) (n int, err error)
	//WriteUint8(n CTypeUint8) error
	//WriteUint16(n CTypeUint16) error
	//WriteUint32(n CTypeUint32) error
}

type buffer struct {
	buf []byte
	off int
}
type row [8]byte

var rowlen = len(row{})

func (b *buffer) empty() bool { return len(b.buf) <= b.off }
func (b *buffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
}

func (b *buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}
func (b *buffer) Write(p []byte) (n int, err error) {
	//b.tryGrowByReslice(len(p))
	return n, err
}
func (b *buffer) ReadRow() (row, error) {
	if b.empty() {
		// Buffer is empty, reset to recover space.
		b.Reset()
		return row{}, io.EOF
	}
	var rowvar row
	n := copy(rowvar[:], b.buf[b.off:])
	if n < len(row{}) {
		return row{}, io.EOF
	}
	b.off += n
	return rowvar, nil
}
func (b *buffer) ReadRows(size int) ([]row, int, error) {
	blocks := size / rowlen
	mod := size % rowlen
	if mod != 0 {
		blocks += 1
	}
	var buf = make([]row, 0)
	for i := 0; i < blocks; i++ {
		in, err := b.ReadRow()
		if err != nil {
			return nil, 0, err
		}
		buf = append(buf, in)
	}
	return buf, mod, nil
}

func (b *buffer) ReadUint8() (CTypeUint8, error) {
	r, err := b.ReadRow()
	if err != nil {
		return 0, err
	}
	return CTypeUint8(r[0]), nil
}

func (b *buffer) ReadUint16() (CTypeUint16, error) {
	r, err := b.ReadRow()
	if err != nil {
		return 0, err
	}
	return CTypeUint16(binary.LittleEndian.Uint16(r[:])), nil
}

func (b *buffer) ReadUint32() (CTypeUint32, error) {
	r, err := b.ReadRow()
	if err != nil {
		return 0, err
	}
	return CTypeUint32(binary.LittleEndian.Uint32(r[:])), nil
}

func (b *buffer) ReadString(size int) (CTypeString, error) {
	r, n, err := b.ReadRows(size)
	if err != nil {
		return "", err
	}
	buf := make([]byte, len(r)*rowlen)
	for i := 0; i < len(r); i++ {
		start := i * rowlen
		end := (i * rowlen) + rowlen
		copy(buf[start:end], r[i][:])
	}
	return CTypeString(buf[0 : len(buf)-n]), nil
}

func NewBuffer(data []byte) *buffer {
	return &buffer{
		buf: data,
	}
}
