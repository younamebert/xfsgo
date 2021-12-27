package vm

import (
	"math/big"
	"xfsgo/common"
)

type token struct {
	BuiltinContract
	name        CTypeString
	symbol      CTypeString
	decimals    CTypeUint8
	totalSupply CTypeUint256
	balances    map[CTypeAddress]CTypeUint256
}

func (t *token) Create(
	name CTypeString,
	symbol CTypeString,
	decimals CTypeUint8,
	totalSupply CTypeUint256) error {
	t.name = name
	t.symbol = symbol
	t.decimals = decimals
	t.totalSupply = totalSupply
	return nil
}

func (t *token) BuiltinId() uint8 {
	return 0x01
}

func (t *token) Name() CTypeString {
	return t.name
}

func (t *token) Symbol() CTypeString {
	return t.symbol
}

func (t *token) Decimals() CTypeUint8 {
	return t.decimals
}

func (t *token) TotalSupply() CTypeUint256 {
	return t.totalSupply
}
func (t *token) BalanceOf(common.Address) CTypeUint256 {
	return CTypeUint256{}
}
func (t *token) Transfer(common.Address, common.Address) bool {
	return false
}
func (t *token) TransferFrom(common.Address, common.Address, *big.Int) bool {
	return false
}
func (t *token) Approve(common.Address, *big.Int) bool {
	return false
}
func (t *token) Allowance(common.Address, common.Address) bool {
	return false
}
