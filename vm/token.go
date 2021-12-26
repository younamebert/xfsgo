package vm

import (
	"math/big"
	"xfsgo/common"
)

type token struct {
	BuiltinContract
	name        CTypeString
	symbol      CTypeString
	decimals    CTypeUINT8
	totalSupply CTypeUINT256
	balances    map[CTypeAddress]CTypeUINT256
}

func (t *token) Create(
	name CTypeString,
	symbol CTypeString,
	decimals CTypeUINT8,
	totalSupply CTypeUINT256) error {
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

func (t *token) Decimals() CTypeUINT8 {
	return t.decimals
}

func (t *token) TotalSupply() CTypeUINT256 {
	return t.totalSupply
}
func (t *token) BalanceOf(common.Address) CTypeUINT256 {
	return CTypeUINT256{}
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

func BuildTokenCreateCodes(
	name CTypeString,
	symbol CTypeString,
	decimals CTypeUINT8,
	totalSupply CTypeUINT256) []byte {
	return nil
}
