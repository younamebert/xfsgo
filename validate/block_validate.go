package validate

import (
	"fmt"
	"math/big"
	"xfsgo"

	"xfsgo/consensus"

	"xfsgo/params"
)

// BlockValidator is responsible for validating block headers, uncles and
// processed state.
//
// BlockValidator implements Validator.
type BlockValidator struct {
	config *params.ChainConfig // Chain configuration options
	bc     *xfsgo.BlockChain   // Canonical block chain
	engine consensus.Engine    // Consensus engine used for validating
}

// NewBlockValidator returns a new block validator which is safe for re-use
func NewBlockValidator(config *params.ChainConfig, blockchain *xfsgo.BlockChain, engine consensus.Engine) *BlockValidator {
	validator := &BlockValidator{
		config: config,
		engine: engine,
		bc:     blockchain,
	}
	return validator
}

// ValidateBody validates the given block's uncles and verifies the the block
// header's transaction and uncle roots. The headers are assumed to be already
// validated at this point.
func (v *BlockValidator) ValidateBody(block *xfsgo.Block) error {
	// Check whether the block's known, and if not, that it's linkable
	if v.bc.HasBlockAndState(block.HeaderHash()) {
		return ErrKnownBlock
	}

	if !v.bc.HasBlockAndState(block.HashPrevBlock()) {
		return consensus.ErrUnknownAncestor
	}

	// Header validity is known at this point, check the uncles and transactions

	header := block.GetHeader()
	headTxHash := header.TxHash()
	if txhash := xfsgo.CalcTxsRootHash(block.Transactions); txhash != headTxHash {
		return fmt.Errorf("transaction root hash mismatch: have %x, want %x", txhash.Hex(), headTxHash.Hex())
	}
	return nil
}

// ValidateState validates the various changes that happen after a state
// transition, such as amount of used gas, the receipt roots and the state root
// itself. ValidateState returns a database batch if the validation was a success
// otherwise nil and an error is returned.
func (v *BlockValidator) ValidateState(block, parent *xfsgo.Block, receipts []*xfsgo.Receipt, usedGas *big.Int) error {
	header := block.GetHeader()

	if block.GasUsed().Cmp(usedGas) != 0 {
		return fmt.Errorf("invalid gas used (remote: %v local: %v)", block.GasUsed(), usedGas)
	}
	// Validate the received block's bloom with the one derived from the generated receipts.
	// For valid blocks this should always validate to true.
	retsHash := xfsgo.CalcReceiptRootHash(receipts)
	ReceiptsHash := header.ReceiptsHash()
	if retsHash != ReceiptsHash {
		return fmt.Errorf("invalid bloom (remote: %x  local: %x)", ReceiptsHash.Hex(), retsHash.Hex())
	}
	// Validate the state root against the received state root and throw
	// an error if they don't match.
	// if root := statedb.IntermediateRoot(v.config.IsEIP158(header.Number)); header.Root != root {
	// 	return fmt.Errorf("invalid merkle root (remote: %x local: %x)", header.Root, root)
	// }
	return nil
}

func (v *BlockValidator) ValidateDposState(block *xfsgo.Block) error {
	header := block.GetHeader()
	localRoot := block.DposCtx().Root()
	remoteRoot := header.DposContext.Root()
	if remoteRoot != localRoot {
		return fmt.Errorf("invalid dpos root (remote: %x local: %x)", remoteRoot, localRoot)
	}
	return nil
}
