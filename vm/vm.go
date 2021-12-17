package vm

import (
	"errors"
	"xfsgo/common"
	"xfsgo/core"
	"xfsgo/crypto"
)

type VM interface {
	Run(common.Address, []byte, []byte) error
	Create(common.Address, []byte) error
	Call(common.Address, []byte) error
}

const (
	magicXVMCode = uint32(9168)
)
const (
	typeXIP2 = uint8(0x01)
	typeXIP3 = uint8(0x02)
)

var (
	errUnknownContractType = errors.New("unknown contract type")
)

type xvm struct {
	stateTree core.StateTree
}

func NewXVM(st core.StateTree) *xvm {
	return &xvm{
		stateTree: st,
	}
}

type codeHeader struct {
	magic uint32
	mType uint8
}

func readCodeHeader(code []byte) (*codeHeader, error) {

	return nil, nil
}

func (vm *xvm) readPayload(h *codeHeader, code []byte) (payload, error) {
	if h.magic == magicXVMCode {
		switch h.mType {
		case typeXIP2:
		case typeXIP3:
		default:
			return nil, errUnknownContractType
		}
	}
}

func (vm *xvm) Run(addr common.Address, code []byte, input []byte) error {
	header, err := readCodeHeader(code)
	if err != nil {
		return err
	}
	pl, err := vm.readPayload(header, code)
	if err != nil {
		return err
	}
	if input == nil {
		if err = pl.Create(addr); err != nil {
			return err
		}
	} else if err = pl.Call(input); err != nil {
		return err
	}
	return nil
}
func (vm *xvm) Create(addr common.Address, code []byte) error {
	nonce := vm.stateTree.GetNonce(addr)
	caddr := crypto.CreateAddress(addr.Hash(), nonce)
	if err := vm.Run(caddr, code, nil); err != nil {
		return err
	}
	return nil
}

func (vm *xvm) Call(address common.Address, input []byte) error {
	code := vm.stateTree.GetCode(address)
	if err := vm.Run(address, code, input); err != nil {
		return err
	}
	return nil
}
