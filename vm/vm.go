package vm

import (
	"encoding/binary"
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
	magicNumberXVM = uint16(9168)
)

var (
	errUnknownMagicNumber  = errors.New("unknown magic number")
	errUnknownContractType = errors.New("unknown contract type")
)

type xvm struct {
	stateTree core.StateTree
	builtin   map[uint8]BuiltinContract
	returnBuf Buffer
}

func NewXVM(st core.StateTree) *xvm {
	vm := &xvm{
		stateTree: st,
		builtin:   make(map[uint8]BuiltinContract),
		returnBuf: NewBuffer(nil),
	}
	vm.registerBuiltinId(new(token))
	return vm
}

func (vm *xvm) newXVMPayload(contract BuiltinContract, address common.Address) (*xvmPayload, error) {
	return &xvmPayload{
		address:   address,
		stateTree: vm.stateTree,
		contract:  contract,
		resultBuf: vm.returnBuf,
	}, nil
}
func (vm *xvm) createPayload(address common.Address, id uint8) (payload, error) {
	if pk, exists := vm.builtin[id]; exists {
		return vm.newXVMPayload(pk, address)
	}
	return nil, errUnknownContractType
}
func (vm *xvm) registerBuiltinId(b BuiltinContract) {
	if _, exists := vm.builtin[b.BuiltinId()]; !exists {
		vm.builtin[b.BuiltinId()] = b
	}
}
func (vm *xvm) readCode(code []byte, input []byte) (id uint8, err error) {
	if code == nil && input != nil {
		code = make([]byte, 3)
		copy(code[:], input[:])
	}
	if code == nil || len(code) < 3 {
		return 0, errors.New("eof")
	}
	m := binary.LittleEndian.Uint16(code[:2])
	if m != magicNumberXVM {
		return 0, errUnknownMagicNumber
	}
	id = code[2]
	return
}
func (vm *xvm) Run(addr common.Address, code []byte, input []byte) error {
	id, err := vm.readCode(code, input)
	if err != nil {
		return err
	}
	pl, err := vm.createPayload(addr, id)
	if err != nil {
		return err
	}
	if code == nil {
		var realInput = make([]byte, len(input)-3)
		copy(realInput[:], input[3:])
		if err = pl.Create(realInput); err != nil {
			return err
		}
		return nil
	}
	if err = pl.Call(input); err != nil {
		return err
	}
	return nil
}
func (vm *xvm) Create(addr common.Address, input []byte) error {
	nonce := vm.stateTree.GetNonce(addr)
	caddr := crypto.CreateAddress(addr.Hash(), nonce)
	if err := vm.Run(caddr, nil, input); err != nil {
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
