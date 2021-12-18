package vm

import (
	"encoding/hex"
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

func readHex2Bytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
func TestXvm_Create(t *testing.T) {
	st := newTestStateTree()
	vm := NewXVM(st)
	code, err := readHex2Bytes("d0230000014759498ac2a719c619e2c8cf8ee60af2d2407425e95d308eb208425b2a6d427a")
	if err != nil {
		t.Fatal(err)
	}
	if err = vm.Create(common.Address{}, code); err != nil {
		t.Fatal(err)
	}
}

func TestXvm_Run(t *testing.T) {

}
