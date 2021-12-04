package simplevm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	_ = iota
	OpLoad
	OpStore
	OpPush
	OpPop
	OpAdd
)

type vmstack struct {
	list [][]byte
}

func (vstack *vmstack) push(data []byte) {
	var mdata [8]byte
	copy(mdata[:], data)
	vstack.list = append(vstack.list, mdata[:])
}
func (vstack *vmstack) pushUint64(num uint64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], num)
	vstack.push(data[:])
}
func (vstack *vmstack) pushUint32(num uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], num)
	vstack.push(data[:])
}
func (vstack *vmstack) pop() []byte {
	if len(vstack.list) == 0 {
		return nil
	}
	data := vstack.list[len(vstack.list)-1]
	vstack.list = vstack.list[0 : len(vstack.list)-1]
	return data
}

func (vstack *vmstack) popUint32() uint32 {
	data := vstack.pop()
	return binary.LittleEndian.Uint32(data[:4])
}

func (vstack *vmstack) popUint64() uint64 {
	data := vstack.pop()
	return binary.LittleEndian.Uint64(data[:8])
}

type storage interface {
	WriteData([]byte) (int, error)
	ReadData([]byte) (int, error)
}

type SimpleVM struct {
	stack       *vmstack
	dataStorage storage
}

func New(st storage) *SimpleVM {
	return &SimpleVM{
		stack:       new(vmstack),
		dataStorage: st,
	}
}

func readData(reader io.Reader) ([]byte, error) {
	var data [8]byte
	n, err := reader.Read(data[:])
	if err != nil && err == io.EOF {
		return nil, nil
	}
	if n < 8 {
		return nil, errors.New("get data err")
	}
	return data[:], nil
}
func (vm *SimpleVM) Exec(s []byte) error {
	reader := bytes.NewReader(s)
out:
	for {
		var opcodes [4]byte
		n, err := reader.Read(opcodes[:])
		if err != nil && err == io.EOF {
			break out
		} else if n < 4 {
			return errors.New("get op_code err")
		}
		opcode := binary.LittleEndian.Uint32(opcodes[:])
		switch opcode {
		case OpLoad:
			var data [8]byte
			n, err := vm.dataStorage.ReadData(data[:])
			if err != nil {
				return err
			}
			vm.stack.push(data[:n])
		case OpStore:
			item := vm.stack.pop()
			if _, err := vm.dataStorage.WriteData(item); err != nil {
				return err
			}
		case OpPush:
			data, err := readData(reader)
			if err != nil {
				return err
			}
			vm.stack.push(data)
		case OpPop:
			vm.stack.pop()
		case OpAdd:
			a := vm.stack.popUint64()
			b := vm.stack.popUint64()
			vm.stack.pushUint64(a + b)
		default:
			return errors.New("unknown op_code")
		}
	}
	return nil
}
