package xfsgo

// import (
// 	"errors"
// 	"fmt"
// 	"math/big"
// 	"xfsgo/common"
// 	"xfsgo/consensus"
// 	"xfsgo/params"
// )

// // BlockValidator is responsible for validating block headers, uncles and
// // processed state.
// //
// // BlockValidator implements Validator.
// type BlockValidator struct {
// 	config *params.ChainConfig // Chain configuration options
// 	bc     *BlockChain         // Canonical block chain
// 	engine consensus.Engine    // Consensus engine used for validating
// }

// var (
// 	// ErrKnownBlock is returned when a block to import is already known locally.
// 	ErrKnownBlock = errors.New("block already known")

// 	// ErrGasLimitReached is returned by the gas pool if the amount of gas required
// 	// by a transaction is higher than what's left in the block.
// 	ErrGasLimitReached = errors.New("gas limit reached")

// 	// ErrBlacklistedHash is returned if a block to import is on the blacklist.
// 	ErrBlacklistedHash = errors.New("blacklisted hash")

// 	// ErrNonceTooHigh is returned if the nonce of a transaction is higher than the
// 	// next one expected based on the local chain.
// 	ErrNonceTooHigh = errors.New("nonce too high")
// )

// // NewBlockValidator returns a new block validator which is safe for re-use
// func NewBlockValidator(config *params.ChainConfig, blockchain *BlockChain, engine consensus.Engine) *BlockValidator {
// 	validator := &BlockValidator{
// 		config: config,
// 		engine: engine,
// 		bc:     blockchain,
// 	}
// 	return validator
// }

// // ValidateBody validates the given block's uncles and verifies the the block
// // header's transaction and uncle roots. The headers are assumed to be already
// // validated at this point.
// func (v *BlockValidator) ValidateBody(block *Block) error {
// 	header := block.Header
// 	// Check whether the block's known, and if not, that it's linkable
// 	if v.bc.GetBlockByHash(header.HeaderHash()) == nil {
// 		return ErrKnownBlock
// 	}

// 	x := new(big.Float).SetUint64(header.Height)
// 	y := new(big.Float).SetInt64(1)
// 	result := v.bc.GetBlockByNumber(common.BigSub(x, y).Uint64())
// 	HashPrev := result.Header.HashPrevBlock
// 	if header.HashHex() != HashPrev.Hex() {
// 		return consensus.ErrUnknownAncestor
// 	}
// 	return nil
// }

// // ValidateState validates the various changes that happen after a state
// // transition, such as amount of used gas, the receipt roots and the state root
// // itself. ValidateState returns a database batch if the validation was a success
// // otherwise nil and an error is returned.
// func (v *BlockValidator) ValidateState(block, parent *Block, statedb *StateTree, receipts []*Receipt, usedGas *big.Int) error {
// 	header := block.Header
// 	if header.GasUsed.Cmp(usedGas) != 0 {
// 		return fmt.Errorf("invalid gas used (remote: %v local: %v)", block.GasUsed(), usedGas)
// 	}
// 	return nil
// }

// func (v *BlockValidator) ValidateDposState(block *Block) error {
// 	header := block.GetHeader()
// 	localRoot := block.DposCtx().Root()
// 	remoteRoot := header.DposContext.Root()
// 	if remoteRoot != localRoot {
// 		return fmt.Errorf("invalid dpos root (remote: %x local: %x)", remoteRoot, localRoot)
// 	}
// 	return nil
// }

// // CalcGasLimit computes the gas limit of the next block after parent.
// // The result may be modified by the caller.
// // This is miner strategy, not consensus protocol.
// func CalcGasLimit(parent *Block) *big.Int {
// 	// contrib = (parentGasUsed * 3 / 2) / 1024
// 	// contrib := new(big.Int).Mul(parent.GasUsed(), big.NewInt(3))
// 	// contrib = contrib.Div(contrib, big.NewInt(2))
// 	// contrib = contrib.Div(contrib, params.GasLimitBoundDivisor)

// 	// // decay = parentGasLimit / 1024 -1
// 	// decay := new(big.Int).Div(parent.GasLimit(), params.GasLimitBoundDivisor)
// 	// decay.Sub(decay, big.NewInt(1))

// 	/*
// 		strategy: gasLimit of block-to-mine is set based on parent's
// 		gasUsed value.  if parentGasUsed > parentGasLimit * (2/3) then we
// 		increase it, otherwise lower it (or leave it unchanged if it's right
// 		at that usage) the amount increased/decreased depends on how far away
// 		from parentGasLimit * (2/3) parentGasUsed is.
// 	*/
// 	// gl := new(big.Int).Sub(parent.GasLimit(), decay)
// 	// gl = gl.Add(gl, contrib)
// 	// gl.Set(math.BigMax(gl, params.MinGasLimit))

// 	// // however, if we're now below the target (TargetGasLimit) we increase the
// 	// // limit as much as we can (parentGasLimit / 1024 -1)
// 	// if gl.Cmp(params.TargetGasLimit) < 0 {
// 	// 	gl.Add(parent.GasLimit(), decay)
// 	// 	gl.Set(math.BigMin(gl, params.TargetGasLimit))
// 	// }
// 	gl := parent.Header.GasLimit
// 	return gl
// }
