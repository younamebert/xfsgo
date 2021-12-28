package vm

import (
	"encoding/binary"
	"math/big"
	"xfsgo/common"
)

type CTypeUint8 byte
type CTypeBool byte
type CTypeUint16 [2]byte
type CTypeUint32 [4]byte
type CTypeUint64 [8]byte
type CTypeUint256 [32]byte
type CTypeString []byte
type CTypeAddress [25]byte

func (t CTypeUint8) uint8() uint8 {
	return uint8(t)
}

func (t CTypeUint16) uint16() uint16 {
	return binary.LittleEndian.Uint16(t[:])
}

func (t CTypeUint32) uint32() uint32 {
	return binary.LittleEndian.Uint32(t[:])
}

func (t CTypeUint64) uint64() uint64 {
	return binary.LittleEndian.Uint64(t[:])
}

func (t CTypeUint256) bigInt() *big.Int {
	return new(big.Int).SetBytes(t[:])
}

func (t CTypeString) string() string {
	return string(t)
}

func (t CTypeAddress) address() common.Address {
	return common.Bytes2Address(t[:])
}

func (t CTypeBool) bool() bool {
	if t == 1 {
		return true
	}
	return false
}

func newUint8(n uint8) CTypeUint8 {
	return CTypeUint8(n)
}

func newUint16(n uint16) (m CTypeUint16) {
	binary.LittleEndian.PutUint16(m[:], n)
	return
}

func newUint32(n uint32) (m CTypeUint32) {
	binary.LittleEndian.PutUint32(m[:], n)
	return
}
func newUint64(n uint64) (m CTypeUint64) {
	binary.LittleEndian.PutUint64(m[:], n)
	return
}

func newUint256(n *big.Int) (m CTypeUint256) {
	bs := n.Bytes()
	copy(m[:], bs)
	return
}
