// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package xfsgo

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"xfsgo/common"
	"xfsgo/crypto"
	"xfsgo/params"
	"xfsgo/types"
	"xfsgo/xfsvm/vm"
)

var emptyCodeHash = crypto.Keccak256Hash(nil)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	gasFeeCap  *big.Int
	gasTipCap  *big.Int
	initialGas uint64
	value      *big.Int
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	To() common.Address
	GasPrice() *big.Int
	Gas() uint64
	Value() *big.Int
	Nonce() uint64
	Data() []byte
	AccessList() types.AccessList
}

// ExecutionResult includes all output after executing given evm
// message no matter the execution itself is successful or not.
type ExecutionResult struct {
	UsedGas    uint64 // Total used gas but include the refunded gas
	Err        error  // Any error encountered during the execution(listed in core/vm/errors.go)
	ReturnData []byte // Returned data from evm(function result or data supplied with revert opcode)
}

// Unwrap returns the internal evm error which allows us for further
// analysis outside.
func (result *ExecutionResult) Unwrap() error {
	return result.Err
}

// Failed returns the indicator whether the execution is successful or not
func (result *ExecutionResult) Failed() bool { return result.Err != nil }

// Return is a helper function to help caller distinguish between revert reason
// and function return. Return returns the data after execution if no error occurs.
func (result *ExecutionResult) Return() []byte {
	if result.Err != nil {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// Revert returns the concrete revert reason if the execution is aborted by `REVERT`
// opcode. Note the reason can be nil if no data supplied with revert opcode.
func (result *ExecutionResult) Revert() []byte {
	if result.Err != vm.ErrExecutionReverted {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, accessList types.AccessList, isContractCreation bool, isHomestead, isEIP2028 bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if isContractCreation && isHomestead {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		nonZeroGas := params.TxDataNonZeroGasFrontier
		if isEIP2028 {
			nonZeroGas = params.TxDataNonZeroGasEIP2028
		}
		if (math.MaxUint64-gas)/nonZeroGas < nz {
			return 0, errors.New("gas uint64 overflow")
		}
		gas += nz * nonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, errors.New("gas uint64 overflow")
		}
		gas += z * params.TxDataZeroGas
	}
	if accessList != nil {
		gas += uint64(len(accessList)) * params.TxAccessListAddressGas
		gas += uint64(accessList.StorageKeys()) * params.TxAccessListStorageKeyGas
	}
	return gas, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: msg.GasPrice(),
		value:    msg.Value(),
		data:     msg.Data(),
		state:    evm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) (*ExecutionResult, error) {
	return NewStateTransition(evm, msg, gp).TransitionDb()
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || (st.msg.To() == common.Address{}) {
		return common.Address{}
	}
	return st.msg.To()
}

// func txPreCheck(stateTree *StateTree, tx *Transaction, gp *GasPool, gas *big.Int) (*StateObj, error) {
// 	fromaddr, err := tx.FromAddr()
// 	if err != nil {
// 		return nil, err
// 	}
// 	sender := stateTree.GetOrNewStateObj(fromaddr)
// 	if sender.GetNonce() != tx.Nonce {
// 		return sender, fmt.Errorf("nonce err: want=%d, got=%d", sender.GetNonce(), tx.Nonce)
// 	}
// 	if err = buyGas(sender, tx, gp, gas); err != nil {
// 		return sender, err
// 	}
// 	return sender, nil
// }

// func useGas(gas, amount *big.Int) error {
// 	if gas.Cmp(amount) < 0 {
// 		return errors.New("out of gas")
// 	}
// 	gas.Sub(gas, amount)
// 	return nil
// }

// func buyGas(sender *StateObj, tx *Transaction, gp *GasPool, gas *big.Int) error {
// 	mgval := new(big.Int).Mul(tx.GasPrice, tx.GasLimit)
// 	if sender.GetBalance().Cmp(mgval) < 0 {
// 		return fmt.Errorf("per-buy gas err, balance is not enough")
// 	}
// 	if err := gp.SubGas(tx.GasLimit); err != nil {
// 		//logrus.Warnf("gas limit out: %s, gp=%s", tx.GasLimit, gp)
// 		return err
// 	}
// 	gas.Add(gas, tx.GasLimit)
// 	sender.SubBalance(mgval)
// 	return nil
// }

func (st *StateTransition) buyGas() error {
	// mgval := new(big.Int).SetUint64(st.msg.Gas())
	// mgval = mgval.Mul(mgval, st.gasPrice)
	// balanceCheck := mgval
	// if st.gasFeeCap != nil {
	// 	balanceCheck = new(big.Int).SetUint64(st.msg.Gas())
	// 	balanceCheck = balanceCheck.Mul(balanceCheck, st.gasFeeCap)
	// 	balanceCheck.Add(balanceCheck, st.value)
	// }
	// if have, want := st.state.GetBalance(st.msg.From()), balanceCheck; have.Cmp(want) < 0 {
	// 	return fmt.Errorf("%w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From().Hex(), have, want)
	// }
	// if err := st.gp.SubGas(st.msg.Gas()); err != nil {
	// 	return err
	// }
	// st.gas += st.msg.Gas()

	// st.initialGas = st.msg.Gas()
	// st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) refundGas(refundQuotient uint64) {
	// Apply refund counter, capped to a refund quotient
	refund := st.gasUsed() / refundQuotient
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return ETH for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(new(big.Int).SetUint64(st.gas))
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}

func (st *StateTransition) txPreCheck() error {

	fromaddr := st.msg.From()

	stNonce := st.state.GetNonce(fromaddr)
	if stNonce != st.msg.Nonce() {
		return fmt.Errorf("nonce err: want=%d, got=%d", stNonce, st.msg.Nonce())
	}

	return st.buyGas()
}

// TransitionDb will transition the state by applying the current message and
// returning the evm execution result with following fields.
//
// - used gas:
//      total gas used (including gas being refunded)
// - returndata:
//      the returned data from evm
// - concrete execution error:
//      various **EVM** error which aborts the execution,
//      e.g. ErrOutOfGas, ErrExecutionReverted
//
// However if any consensus issue encountered, return the error directly with
// nil evm execution result.
func (st *StateTransition) TransitionDb() (*ExecutionResult, error) {
	// First check this message satisfies all consensus rules before
	// applying the message. The rules include these clauses
	//
	// 1. the nonce of the message caller is correct
	// 2. caller has enough balance to cover transaction fee(gaslimit * gasprice)
	// 3. the amount of gas required is available in the block
	// 4. the purchased gas is enough to cover intrinsic usage
	// 5. there is no overflow when calculating intrinsic gas
	// 6. caller has enough balance to cover asset transfer for **topmost** call

	// Check clauses 1-3, buy gas if everything is correct
	if err := st.txPreCheck(); err != nil {
		return nil, err
	}
	msg := st.msg
	sender := vm.AccountRef(msg.From())
	homestead := st.evm.ChainConfig().IsHomestead(st.evm.Context.BlockNumber)
	istanbul := st.evm.ChainConfig().IsIstanbul(st.evm.Context.BlockNumber)
	london := st.evm.ChainConfig().IsLondon(st.evm.Context.BlockNumber)
	contractCreation := msg.To() == common.Address{}

	// Check clauses 4-5, subtract intrinsic gas if everything is correct
	gas, err := IntrinsicGas(st.data, st.msg.AccessList(), contractCreation, homestead, istanbul)
	if err != nil {
		return nil, err
	}
	if st.gas < gas {
		return nil, fmt.Errorf("%w: have %d, want %d", errors.New("intrinsic gas too low"), st.gas, gas)
	}
	st.gas -= gas

	// Check clause 6
	if msg.Value().Sign() > 0 && !st.evm.Context.CanTransfer(st.state, msg.From(), msg.Value()) {
		return nil, fmt.Errorf("%w: address %v", errors.New("insufficient funds for transfer"), msg.From())
	}

	// Set up the initial access list.
	// if rules := st.evm.ChainConfig().Rules(st.evm.Context.BlockNumber); rules.IsBerlin {
	// 	st.state.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())
	// }
	var (
		ret   []byte
		vmerr error // vm errors do not effect consensus and are therefore not assigned to err
	)
	if contractCreation {
		ret, _, st.gas, vmerr = st.evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = st.evm.Call(sender, st.to(), st.data, st.gas, st.value)
	}

	if !london {
		// Before EIP-3529: refunds were capped to gasUsed / 2
		st.refundGas(params.RefundQuotient)
	} else {
		// After EIP-3529: refunds are capped to gasUsed / 5
		st.refundGas(params.RefundQuotientEIP3529)
	}
	// effectiveTip := st.gasPrice
	// if london {
	// 	effectiveTip = cmath.BigMin(st.gasTipCap, new(big.Int).Sub(st.gasFeeCap, st.evm.Context.BaseFee))
	// }
	// st.state.AddBalance(st.evm.Context.Coinbase, new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), effectiveTip))

	return &ExecutionResult{
		UsedGas:    st.gasUsed(),
		Err:        vmerr,
		ReturnData: ret,
	}, nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(bc ChainContext, author *common.Address, gp *GasPool, statedb *StateTree, header *BlockHeader, tx *Transaction, usedGas *uint64, cfg vm.Config) (*Receipt, error) {
	msg, err := tx.AsMessage()
	if err != nil {
		return nil, err
	}
	// Create a new context to be used in the EVM environment
	blockContext := NewEVMBlockContext(header, bc, author)
	vmenv := vm.NewEVM(blockContext, vm.TxContext{}, statedb, cfg)
	return applyTransaction(msg, bc, author, gp, statedb, new(big.Int).SetUint64(header.Height), header.HeaderHash(), tx, usedGas, vmenv)
}

func applyTransaction(msg MessageImp, bc ChainContext, author *common.Address, gp *GasPool, statedb *StateTree, blockNumber *big.Int, blockHash common.Hash, tx *Transaction, usedGas *uint64, evm *vm.EVM) (*Receipt, error) {
	// Create a new context to be used in the EVM environment.
	txContext := NewEVMTxContext(msg)
	evm.Reset(txContext, statedb)

	// Apply the transaction to the current state (included in the env).
	_, err := ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, err
	}

	// Update the state with pending changes.
	// var root []byte
	// if config.IsByzantium(blockNumber) {
	// 	statedb.Finalise(true)
	// } else {
	// 	root = statedb.IntermediateRoot(config.IsEIP158(blockNumber)).Bytes()
	// }
	// *usedGas += result.UsedGas

	// // Create a new receipt for the transaction, storing the intermediate root and gas used
	// // by the tx.
	// receipt := &types.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: *usedGas}
	// if result.Failed() {
	// 	receipt.Status = types.ReceiptStatusFailed
	// } else {
	// 	receipt.Status = types.ReceiptStatusSuccessful
	// }
	// receipt.TxHash = tx.Hash()
	// receipt.GasUsed = result.UsedGas

	// // If the transaction created a contract, store the creation address in the receipt.
	// if msg.To() == nil {
	// 	receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	// }

	// // Set the receipt logs and create the bloom filter.
	// receipt.Logs = statedb.GetLogs(tx.Hash(), blockHash)
	// receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	// receipt.BlockHash = blockHash
	// receipt.BlockNumber = blockNumber
	// receipt.TransactionIndex = uint(statedb.TxIndex())
	return nil, err
}
