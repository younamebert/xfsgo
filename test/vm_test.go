// Copyright 2016 The go-ethereum Authors

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

package test

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/params"
	"xfsgo/xfsvm"

	"xfsgo/xfsvm/abi"
	"xfsgo/xfsvm/vm"
)

var (
	testHash    = common.Hash{}
	fromAddress = common.StrB58ToAddress("bQMEGRNy9iAyfEugoBoVnHpLnuprV381H")
	toAddress   = common.Hash{} //common.StrB58ToAddress("fLmaBWVaCjwGJn1NJomQ9SzpcfPm18kre")
	amount      = big.NewInt(1)
	nonce       = uint64(0)
	gasLimit    = big.NewInt(100000)
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func loadBin(str1 string) []byte {
	str := "0x"
	str += str1

	addr := common.StrB58ToAddress(str)
	return addr.Bytes()

}
func loadAbi(filename string) abi.ABI {
	abiFile, err := os.Open(filename)
	must(err)
	defer abiFile.Close()
	abiObj, err := abi.JSON(abiFile)
	must(err)
	return abiObj
}

type ChainContext struct{}

func (cc ChainContext) GetHeader(hash common.Hash, number uint64) *xfsgo.BlockHeader {
	return &xfsgo.BlockHeader{
		Height:        0,
		Version:       0,
		HashPrevBlock: hash,
		Timestamp:     0,
		Coinbase:      fromAddress,
		// merkle tree root hash
		StateRoot:        common.Hash{},
		TransactionsRoot: common.Hash{},
		ReceiptsRoot:     common.Hash{},
		GasLimit:         big.NewInt(100000),
		GasUsed:          big.NewInt(100000),
		// pow consensus.
		Bits:       4278190109,
		Nonce:      0,
		ExtraNonce: 0,
	}
}

func TestNewEVMBlockContext(t *testing.T) {

	fmt.Println(fromAddress, toAddress)
	abiFileName := "../xfsvm/file/coin_sol_Coin.abi"
	// binFileName := "../xfsvm/file/coin_sol_Coin.bin"
	data := "PUSH1 0x80 PUSH1 0x40 MSTORE CALLVALUE DUP1 ISZERO PUSH2 0x10 JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH2 0x150 DUP1 PUSH2 0x20 PUSH1 0x0 CODECOPY PUSH1 0x0 RETURN INVALID PUSH1 0x80 PUSH1 0x40 MSTORE CALLVALUE DUP1 ISZERO PUSH2 0x10 JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH1 0x4 CALLDATASIZE LT PUSH2 0x36 JUMPI PUSH1 0x0 CALLDATALOAD PUSH1 0xE0 SHR DUP1 PUSH4 0x2E64CEC1 EQ PUSH2 0x3B JUMPI DUP1 PUSH4 0x6057361D EQ PUSH2 0x59 JUMPI JUMPDEST PUSH1 0x0 DUP1 REVERT JUMPDEST PUSH2 0x43 PUSH2 0x75 JUMP JUMPDEST PUSH1 0x40 MLOAD PUSH2 0x50 SWAP2 SWAP1 PUSH2 0xD9 JUMP JUMPDEST PUSH1 0x40 MLOAD DUP1 SWAP2 SUB SWAP1 RETURN JUMPDEST PUSH2 0x73 PUSH1 0x4 DUP1 CALLDATASIZE SUB DUP2 ADD SWAP1 PUSH2 0x6E SWAP2 SWAP1 PUSH2 0x9D JUMP JUMPDEST PUSH2 0x7E JUMP JUMPDEST STOP JUMPDEST PUSH1 0x0 DUP1 SLOAD SWAP1 POP SWAP1 JUMP JUMPDEST DUP1 PUSH1 0x0 DUP2 SWAP1 SSTORE POP POP JUMP JUMPDEST PUSH1 0x0 DUP2 CALLDATALOAD SWAP1 POP PUSH2 0x97 DUP2 PUSH2 0x103 JUMP JUMPDEST SWAP3 SWAP2 POP POP JUMP JUMPDEST PUSH1 0x0 PUSH1 0x20 DUP3 DUP5 SUB SLT ISZERO PUSH2 0xB3 JUMPI PUSH2 0xB2 PUSH2 0xFE JUMP JUMPDEST JUMPDEST PUSH1 0x0 PUSH2 0xC1 DUP5 DUP3 DUP6 ADD PUSH2 0x88 JUMP JUMPDEST SWAP2 POP POP SWAP3 SWAP2 POP POP JUMP JUMPDEST PUSH2 0xD3 DUP2 PUSH2 0xF4 JUMP JUMPDEST DUP3 MSTORE POP POP JUMP JUMPDEST PUSH1 0x0 PUSH1 0x20 DUP3 ADD SWAP1 POP PUSH2 0xEE PUSH1 0x0 DUP4 ADD DUP5 PUSH2 0xCA JUMP JUMPDEST SWAP3 SWAP2 POP POP JUMP JUMPDEST PUSH1 0x0 DUP2 SWAP1 POP SWAP2 SWAP1 POP JUMP JUMPDEST PUSH1 0x0 DUP1 REVERT JUMPDEST PUSH2 0x10C DUP2 PUSH2 0xF4 JUMP JUMPDEST DUP2 EQ PUSH2 0x117 JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP JUMP INVALID LOG2 PUSH5 0x6970667358 0x22 SLT KECCAK256 BLOCKHASH 0x4E CALLDATACOPY DELEGATECALL DUP8 0xA8 SWAP11 SWAP4 0x2D 0xCA 0x5E PUSH24 0xFAAF6CA2DE3B991F93D230604B1B8DAAEF64766264736F6C PUSH4 0x43000807 STOP CALLER "

	// data := []byte{byte(vm.PUSH1), 0, byte(vm.PUSH1), 0, byte(vm.RETURN)}
	cc := ChainContext{}
	ctx := xfsvm.NewEVMBlockContext(cc.GetHeader(testHash, 0), cc, &fromAddress)

	msg := xfsvm.NewMessage(fromAddress, nil, nonce, amount, gasLimit, big.NewInt(1), common.Hex2bytes(data), false)
	txContext := xfsvm.NewEVMTxContext(msg)

	stateDB := NewMemStorage()
	stateTree := xfsgo.NewStateTree(stateDB, nil)
	stateTree.GetOrNewStateObj(fromAddress)
	// stateTree.GetOrNewStateObj(toAddress)
	stateTree.AddBalance(fromAddress, big.NewInt(1e18))
	testBalance := stateTree.GetBalance(fromAddress)
	fmt.Println("init testBalance =", testBalance)

	config := params.MainnetChainConfig
	logConfig := vm.LogConfig{}
	structLogger := vm.NewStructLogger(&logConfig)
	vmConfig := vm.Config{Debug: true, Tracer: structLogger /*, JumpTable: vm.NewByzantiumInstructionSet()*/}

	evm := vm.NewEVM(ctx, txContext, stateTree, config, vmConfig)
	contractRef := vm.AccountRef(fromAddress)
	fmt.Println(fromAddress, contractRef)
	contractCode, contractAddr, gasLeftover, err := evm.Create(contractRef, common.Hex2bytes(data), 1000000, big.NewInt(1000000000))
	fmt.Printf("err:%v\n", err)

	fmt.Printf("contractAddr:%v,gasLeftover:%v\n", contractAddr, gasLeftover)
	fmt.Printf("contractCode:%v,code:%v\n", contractCode, stateTree.GetCode(contractAddr))

	testBalance = stateTree.GetBalance(fromAddress)
	contractAddrBal := stateTree.GetBalance(contractAddr)
	fmt.Printf("after create contract, testBalance%v,contractAddr=%v\n", testBalance, contractAddrBal)

	abiObj := loadAbi(abiFileName)
	input, err := abiObj.Pack("store", big.NewInt(1))
	must(err)
	outputs, gasLeftover, vmerr := evm.Call(contractRef, contractAddr, input, stateTree.GetBalance(fromAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	fmt.Printf("call outputs %v\n", outputs)
	fmt.Printf("call contractRef %v\n", contractRef)
	fmt.Printf("call gasLeftover %v\n", gasLeftover)

	sender := common.Bytes2Address(outputs)

	senderAcc := vm.AccountRef(sender)

	input, _ = abiObj.Pack("retrieve")
	outputs, gasLeftover, _ = evm.Call(senderAcc, contractAddr, input, stateTree.GetBalance(fromAddress).Uint64(), big.NewInt(0))

	fmt.Printf("call outputs %v\n", outputs)
	fmt.Printf("call contractRef %v\n", contractRef)
	fmt.Printf("call gasLeftover %v\n", gasLeftover)
}
