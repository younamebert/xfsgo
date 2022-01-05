package test

import (
	"math/big"
	"xfsgo/common"
	"xfsgo/params"
)

var (
	TestChainConfig = &params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      new(big.Int),
		DAOForkBlock:        new(big.Int),
		DAOForkSupport:      false,
		EIP150Block:         new(big.Int),
		EIP150Hash:          common.Hash{},
		EIP155Block:         new(big.Int),
		EIP158Block:         new(big.Int),
		ByzantiumBlock:      new(big.Int),
		ConstantinopleBlock: new(big.Int),
		PetersburgBlock:     new(big.Int),
		IstanbulBlock:       new(big.Int),
		MuirGlacierBlock:    new(big.Int),
		BerlinBlock:         new(big.Int),
		LondonBlock:         new(big.Int),
		Dpos:                &params.DposConfig{},
	}
	// TestChainConfig = &params.DposConfig{
	// 	Validators: common.B58ToAddress([]byte),
	// }
	// Testdpos = dpos.New(chainConfig.Dpos, config.ChainDB)
)

// func Test()
