package vm

import (
	"math/big"
	"xfsgo/common"
	"xfsgo/common/ahash"
	"xfsgo/core"
)

type payload interface {
	Create(address common.Address) error
	Call([]byte) error
}

type tokenPayload struct {
	stateTree   core.StateTree
	name        string
	symbol      string
	decimals    uint8
	totalSupply *big.Int
}

func makeKey(k []byte) (key [32]byte) {
	keyhash := ahash.SHA256(k)
	copy(key[:], keyhash)
	return
}

func newTokenPayload(payload []byte) (*tokenPayload, error) {

	return nil, nil
}
func (p *tokenPayload) readCreateParams() {

}
func (p *tokenPayload) Create(address common.Address) error {
	p.stateTree.SetState(address, makeKey(), "")
	return nil
}
