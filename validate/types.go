package validate

import (
	"math/big"
	"xfsgo"
)

// Validator is an interface which defines the standard for block validation. It
// is only responsible for validating block contents, as the header validation is
// done by the specific consensus engines.
//
type Validator interface {
	// ValidateBody validates the given block's content.
	ValidateBody(block *xfsgo.Block) error

	// ValidateState validates the given statedb and optionally the receipts and
	// gas used.
	ValidateState(block, parent *xfsgo.Block, receipts []*xfsgo.Receipt, usedGas *big.Int) error
	// ValidateDposState validates the given dpos state
	ValidateDposState(block *xfsgo.Block) error
}

// // Processor is an interface for processing blocks using a given initial state.
// //
// // Process takes the block to be processed and the statedb upon which the
// // initial state is based. It should return the receipts generated, amount
// // of gas used in the process and return an error if any of the internal rules
// // failed.
// type Processor interface {
// 	Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, *big.Int, error)
// }