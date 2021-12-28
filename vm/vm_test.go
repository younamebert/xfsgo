package vm

import (
	"bytes"
	"encoding/binary"
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

func readBytes4hex(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
func mustReadBytes4hex(s string) []byte {
	bs, err := readBytes4hex(s)
	if err != nil {
		panic(err)
	}
	return bs
}
func writeStringParams(w Buffer, s CTypeString) {
	slen := len(s)
	var slenbuf [8]byte
	binary.LittleEndian.PutUint64(slenbuf[:], uint64(slen))
	_, _ = w.Write(slenbuf[:])
	_, _ = w.Write(s)
}

var (
	tokenCode = []byte{
		0xd0, 0x23, 0x01,
	}
	tokenCreateFnHash = mustReadBytes4hex("4759498ac2a719c619e2c8cf8ee60af2d2407425e95d308eb208425b2a6d427a")
	tokenCreateParams = func(
		name CTypeString,
		symbol CTypeString,
		decimals CTypeUint8,
		totalSupply CTypeUint256) (d []byte) {
		buf := NewBuffer(nil)
		writeStringParams(buf, name)
		writeStringParams(buf, symbol)
		_, _ = buf.Write([]byte{byte(decimals)})
		_, _ = buf.Write(totalSupply[:])
		return buf.Bytes()
	}
	testAbTokenCreateParams = tokenCreateParams(
		CTypeString("AbCoin"),
		CTypeString("AB"),
		18,
		newUint256(new(big.Int).SetInt64(100)))
)

func TestXvm_Create(t *testing.T) {
	st := newTestStateTree()
	vm := NewXVM(st)
	inputBuf := bytes.NewBuffer(nil)
	inputBuf.Write(tokenCode)
	inputBuf.Write(tokenCreateFnHash)
	inputBuf.Write(testAbTokenCreateParams)
	if err := vm.Create(common.Address{}, inputBuf.Bytes()); err != nil {
		t.Fatal(err)
	}
}

func TestXvm_Run(t *testing.T) {

}
