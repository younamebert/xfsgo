package vm

import (
	"math/big"
	"xfsgo/common"
)

type Token interface {
	GetName() CTypeString
	GetSymbol() CTypeString
	GetDecimals() CTypeUint8
	GetTotalSupply() CTypeUint256
	BalanceOf(address CTypeAddress) CTypeUint256
}

type token struct {
	BuiltinContract
	Name        CTypeString                   `contract:"storage"`
	Symbol      CTypeString                   `contract:"storage"`
	Decimals    CTypeUint8                    `contract:"storage"`
	TotalSupply CTypeUint256                  `contract:"storage"`
	Balances    map[CTypeAddress]CTypeUint256 `contract:"storage"`
}

func (t *token) Create(
	name CTypeString,
	symbol CTypeString,
	decimals CTypeUint8,
	totalSupply CTypeUint256) error {
	t.Name = name
	t.Symbol = symbol
	t.Decimals = decimals
	t.TotalSupply = totalSupply
	t.Balances = make(map[CTypeAddress]CTypeUint256)
	return nil
}

//func (t *token) Constructor(func()) error {
//	t.Name = name
//	t.Symbol = symbol
//	t.Decimals = decimals
//	t.TotalSupply = totalSupply
//	return nil
//}

func (t *token) BuiltinId() uint8 {
	return 0x01
}

func (t *token) GetName() CTypeString {
	return t.Name
}

func (t *token) GetSymbol() CTypeString {
	return t.Symbol
}

func (t *token) GetDecimals() CTypeUint8 {
	return t.Decimals
}

func (t *token) GetTotalSupply() CTypeUint256 {
	return t.TotalSupply
}
func (t *token) BalanceOf(CTypeAddress) CTypeUint256 {
	return CTypeUint256{}
}
func (t *token) Transfer(addr common.Address, val CTypeUint256) CTypeBool {
	return CTypeBool{}
}
func (t *token) TransferFrom(common.Address, common.Address, *big.Int) CTypeBool {
	return CTypeBool{}
}
func (t *token) Approve(common.Address, *big.Int) CTypeBool {
	return CTypeBool{}
}
func (t *token) Allowance(common.Address, common.Address) CTypeBool {
	return CTypeBool{}
}
