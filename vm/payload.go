package vm

import (
	"bytes"
	"errors"
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
	errNotfoundMethod   = errors.New("notfound method")
)

type xvmPayload struct {
	createFn  common.Hash
	stateTree core.StateTree
	address   common.Address
	contract  BuiltinContract
}

func (p *xvmPayload) Create(input []byte) error {
	return p.callFn(p.createFn, input)
}

func (p *xvmPayload) call(fn reflect.Method, input []byte) error {
	buf := NewBuffer(input)
	mType := fn.Type
	n := mType.NumIn()

	var args = make([]reflect.Value, 0)
	for i := 1; i < n; i++ {
		parameterType := mType.In(i)
		switch parameterType.Name() {
		case CTypeStringN:
			ssize, err := buf.ReadUint32()
			if err != nil {
				return err
			}
			s, err := buf.ReadString(int(ssize))
			if err != nil {
				return err
			}
			args = append(args, reflect.ValueOf(s))
		case CTypeUint8N:

		}
	}

	return nil
}
func (p *xvmPayload) callFn(fn common.Hash, input []byte) error {
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
	if m, ok := findMethod(fn); ok {
		if err := p.call(m, input); err != nil {
			return err
		}
	}
	return errNotfoundMethod
}

func (p *xvmPayload) Call([]byte) error {
	return nil
}

func makeKey(k []byte) (key [32]byte) {
	keyhash := ahash.SHA256(k)
	copy(key[:], keyhash)
	return
}
