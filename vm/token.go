package vm

import (
	"math/big"
	"xfsgo/common"
)

type token struct {
	BuiltinContract
	Name        CTypeString                   `contract:"storage" json:"name"`
	Symbol      CTypeString                   `contract:"storage" json:"symbol"`
	Decimals    CTypeUint8                    `contract:"storage" json:"decimals"`
	TotalSupply CTypeUint256                  `contract:"storage" json:"totalSupply"`
	Balances    map[CTypeAddress]CTypeUint256 `contract:"storage" json:"balances"`
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
	t.Balances[CTypeAddress{0x01}] = CTypeUint256{0x02}
	return nil
}

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
func (t *token) BalanceOf(common.Address) CTypeUint256 {
	return CTypeUint256{}
}
func (t *token) Transfer(addr common.Address, val CTypeUint256) CTypeBool {
	return CTypeBool(0)
}
func (t *token) TransferFrom(common.Address, common.Address, *big.Int) CTypeBool {
	return CTypeBool(0)
}
func (t *token) Approve(common.Address, *big.Int) CTypeBool {
	return CTypeBool(0)
}
func (t *token) Allowance(common.Address, common.Address) CTypeBool {
	return CTypeBool(0)
}
