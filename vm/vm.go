package vm

import "xfsgo/common"

type VM interface {
	Create(common.Address, []byte) error
	Call(common.Address, []byte) error
}
type xvm struct {
}

func NewXVM() *xvm {
	return &xvm{}
}

func (vm *xvm) Create(common.Address, []byte) error {
	return nil
}

func (vm *xvm) Call(common.Address, []byte) error {
	return nil
}
