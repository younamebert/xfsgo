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
)

type xvmPayload struct {
	createFn  common.Hash
	stateTree core.StateTree
	address   common.Address
	contract  BuiltinContract
}

func (p *xvmPayload) callCreate() error {
	ct := reflect.TypeOf(p.contract)
	findMethod := func(hash common.Hash) (reflect.Method, bool) {
		for i := 0; i < ct.NumMethod(); i++ {
			sf := ct.Method(i)
			aname := sf.Name
			namehash := ahash.SHA256([]byte(aname))
			if bytes.Equal(hash[:], namehash) {
				return sf, true
			}
		}
		return reflect.Method{}, false
	}
	if m, ok := findMethod(p.createFn); ok {
		m.Type.Elem()
	}
	return errNotfoundCreateFn
}
func (p *xvmPayload) Create([]byte) error {
	//p.callCreate()
	return nil
}

func (p *xvmPayload) Call([]byte) error {
	return nil
}

func makeKey(k []byte) (key [32]byte) {
	keyhash := ahash.SHA256(k)
	copy(key[:], keyhash)
	return
}
