package vm

import (
	"math/big"
	"testing"
	"xfsgo/common"
)

type testStateTree struct {
}

func (t *testStateTree) GetNonce(common.Address) uint64 {
	return 0
}
func (t *testStateTree) GetBalance(common.Address) *big.Int {
	return nil
}
func (t *testStateTree) GetCode(common.Address) []byte {
	return nil
}
func (t *testStateTree) SetState(common.Address, [32]byte, []byte) {

}
func (t *testStateTree) GetStateValue(common.Address, [32]byte) []byte {
	return nil
}
func newTestStateTree() *testStateTree {
	return &testStateTree{}
}
func TestXvm_Create(t *testing.T) {
	st := newTestStateTree()
	vm := NewXVM(st)
	code := make([]byte, 0)
	if err := vm.Create(common.Address{}, code); err != nil {
		t.Fatal(err)
	}
}

func TestXvm_Run(t *testing.T) {

}
