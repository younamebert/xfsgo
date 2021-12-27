// Copyright 2018 The xfsgo Authors
// This file is part of the xfsgo library.
//
// The xfsgo library is free software: you can redistribute it and/or modify
// it under the terms of the MIT Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The xfsgo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// MIT Lesser General Public License for more details.
//
// You should have received a copy of the MIT Lesser General Public License
// along with the xfsgo library. If not, see <https://mit-license.org/>.

package common

import (
	"math/big"
)

var (
	Big0          = new(big.Int).SetInt64(0)
	Big1          = new(big.Int).SetInt64(1)
	Big2          = new(big.Int).SetInt64(2)
	Big10         = new(big.Int).SetInt64(10)
	Big32         = new(big.Int).SetInt64(32)
	Big50         = new(big.Int).SetInt64(50)
	Big64         = new(big.Int).SetInt64(64)
	Big100        = new(big.Int).SetInt64(100)
	Big256        = new(big.Int).SetInt64(256)
	Big32Bits     = new(big.Int).Exp(Big2, Big32, nil)
	Big64Bits     = new(big.Int).Exp(Big2, Big64, nil)
	Big256Bits    = new(big.Int).Exp(Big2, Big256, nil)
	BigMaxUint32  = new(big.Int).Sub(Big32Bits, Big1)
	BigMaxUint64  = new(big.Int).Sub(Big64Bits, Big1)
	BigMaxUint256 = new(big.Int).Sub(Big256Bits, Big1)
)

func ParseString2BigInt(str string) *big.Int {
	if str == "" {
		return Big0
	}
	num, success := new(big.Int).SetString(str, 0)
	if !success {
		return Big0
	}
	return num
}

func BigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}

func BigMin(x, y *big.Int) *big.Int {
	if x.Cmp(y) > 0 {
		return y
	}

	return x
}

func BigSub(x, y *big.Float) *big.Int {
	bigfloat0 := new(big.Float).SetUint64(Big0.Uint64())

	result := new(big.Int)
	info := x.Sub(x, y)
	if info.Cmp(bigfloat0) < 0 {
		info.Set(bigfloat0)
	}
	ais, _ := info.Uint64()
	return result.SetUint64(ais)
}

func BigSubN(x, y uint64) *big.Int {
	xf := new(big.Float).SetUint64(x)
	yf := new(big.Float).SetUint64(y)
	return BigSub(xf, yf)
}

func IsForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	x, ok := new(big.Float).SetString(s.String())
	if !ok {
		return false
	}
	y, ok := new(big.Float).SetString(head.String())
	if !ok {
		return false
	}
	if x.Cmp(y) <= 0 {
		return true
	} else {
		return false
	}
}
