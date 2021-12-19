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
	magicNumberXVM = uint32(9168)
)
const (
	typeXIP2 = uint8(0x01)
	typeXIP3 = uint8(0x02)
)

var (
	errUnknownMagicNumber  = errors.New("unknown magic number")
	errUnknownContractType = errors.New("unknown contract type")
)

type xvm struct {
	stateTree core.StateTree
	builtin   map[uint8]BuiltinContract
}

func NewXVM(st core.StateTree) *xvm {
	vm := &xvm{
		stateTree: st,
		builtin:   make(map[uint8]BuiltinContract),
	}
	vm.registerBuiltinId(new(token))
	return vm
}

func (vm *xvm) newXVMPayload(contract BuiltinContract, address common.Address, ch common.Hash) (*xvmPayload, error) {
	return &xvmPayload{
		createFn:  ch,
		address:   address,
		stateTree: vm.stateTree,
		contract:  contract,
	}, nil
}
func (vm *xvm) readPayload(address common.Address, code []byte) (payload, error) {
	m := binary.LittleEndian.Uint32(code[:4])
	if m != magicNumberXVM {
		return nil, errUnknownMagicNumber
	}
	id := code[4]
	createfnhashbs := code[5 : 5+len(common.Hash{})]
	if pk, exists := vm.builtin[id]; exists {
		return vm.newXVMPayload(pk, address, common.Bytes2Hash(createfnhashbs))
	}
	return nil, errUnknownContractType
}
func (vm *xvm) registerBuiltinId(b BuiltinContract) {
	if _, exists := vm.builtin[b.BuiltinId()]; !exists {
		vm.builtin[b.BuiltinId()] = b
	}
}
func (vm *xvm) Run(addr common.Address, code []byte, input []byte) error {
	pl, err := vm.readPayload(addr, code)
	if err = pl.Create(input); err != nil {
		return err
	}
	return nil
}
func (vm *xvm) Create(addr common.Address, code []byte) error {
	nonce := vm.stateTree.GetNonce(addr)
	caddr := crypto.CreateAddress(addr.Hash(), nonce)
	if err := vm.Run(caddr, code[:37], code[37:]); err != nil {
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
