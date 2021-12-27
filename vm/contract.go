package vm

import (
	"xfsgo/common"
	"xfsgo/core"
)

type ContractPayload interface {
	Create()
}
type BuiltinContract interface {
	GetStateTree() core.StateTree
	GetAddress() common.Address
	Storage([32]byte, []byte)
	StorageFiled(string, []byte)
	BuiltinId() uint8
}
