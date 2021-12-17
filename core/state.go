package core

import (
	"math/big"
	"xfsgo/common"
)

type StateTree interface {
	GetNonce(common.Address) uint64
	GetBalance(common.Address) *big.Int
	GetCode(common.Address) []byte
	SetState(common.Address, [32]byte, []byte)
	GetStateValue(common.Address, [32]byte) []byte
}
