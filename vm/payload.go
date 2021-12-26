package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/core"
)

type payload interface {
	Create([]byte) error
	Call([]byte) error
}

var (
	errNotfoundCreateFn = errors.New("notfound create function")
)

type xvmPayload struct {
	createFn  common.Hash
	stateTree core.StateTree
	address   common.Address
	contract  BuiltinContract
}

func readInputRow(reader io.Reader) (inputRow, error) {
	var row inputRow
	nn, _ := reader.Read(row[:])
	if nn != rowlen {
		return inputRow{}, io.EOF
	}
	return row, nil
}
func readBytesBySize(reader io.Reader, size uint64) ([]byte, error) {
	blocks := size / uint64(rowlen)
	mod := size % uint64(rowlen)
	if mod != 0 {
		blocks += 1
	}
	var buf = make([]byte, blocks*uint64(rowlen))
	for i := uint64(0); i < blocks; i++ {
		in, err := readInputRow(reader)
		if err != nil {
			return nil, err
		}
		start := i * uint64(rowlen)
		end := (i * uint64(rowlen)) + uint64(rowlen)
		copy(buf[start:end], in[:])
	}
	return buf, nil
}
func readStringArg(reader io.Reader) ([]byte, error) {
	nn, err := readInputRow(reader)
	if err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint64(nn[:])
	var buf []byte
	if buf, err = readBytesBySize(reader, size); err != nil {
		return nil, err
	}
	return buf, nil
}
func (p *xvmPayload) Create(input []byte) error {
	ct := reflect.TypeOf(p.contract)
	findMethod := func(hash common.Hash) (reflect.Method, bool) {
		for i := 0; i < ct.NumMethod(); i++ {
			sf := ct.Method(i)
			aname := sf.Name
			namehash := ahash.SHA256([]byte(aname))
			if sf.Type.Kind() == reflect.Func && bytes.Equal(hash[:], namehash) {
				return sf, true
			}
		}
		return reflect.Method{}, false
	}
	buf := bytes.NewBuffer(input)
	if m, ok := findMethod(p.createFn); ok {
		mType := m.Type
		n := mType.NumIn()
		var args = make([][]byte, 0)
		for i := 1; i < n; i++ {
			parameterType := mType.In(i)
			switch parameterType.Name() {
			case CTypeStringN:
				stringbuf, err := readStringArg(buf)
				if err != nil {
					return err
				}
				args = append(args, stringbuf)
			case CTypeUint8N:

			}
		}
		fmt.Printf("%d", len(args))
	}
	return errNotfoundCreateFn
}

func (p *xvmPayload) Call([]byte) error {
	return nil
}

func makeKey(k []byte) (key [32]byte) {
	keyhash := ahash.SHA256(k)
	copy(key[:], keyhash)
	return
}
