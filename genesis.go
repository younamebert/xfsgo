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

package xfsgo

import (
	"fmt"
	"math/big"
	"xfsgo/avlmerkle"
	"xfsgo/common"
	"xfsgo/params"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

var (
	MainNetGenesisBits = uint32(267386909)
	TestNetGenesisBits = uint32(4278190110)
	GenesisBits        = MainNetGenesisBits
)

type GenesisConfig struct {
	StateDB, ChainDB         badger.IStorage
	GenesisCoinbase, Balance string
	Debug                    bool
}

// Genesis specifies the header fields, state of a genesis block. It also defines accounts
type Genesis struct {
	Version       uint32              `json:"version"`
	HashPrevBlock string              `json:"hash_prev_block"`
	Timestamp     string              `json:"timestamp"`
	Coinbase      string              `json:"coinbase"`
	Bits          uint32              `json:"bits"`
	Nonce         uint32              `json:"nonce"`
	ExtraNonce    uint64              `json:"extranonce"`
	Accounts      map[string]string   `json:"accounts"`
	GasLimit      *big.Int            `json:"gas_limit"`
	Config        *params.ChainConfig `json:"config"`
	GenesisConfig
}

// WriteGenesisBlock constructs the genesis block for the blockchain and stores it in the hd.
// func WriteGenesisBlock(stateDB, chainDB badger.IStorage, genesis *Genesis) (*Block, error) {
// 	return WriteGenesisBlockN(stateDB, chainDB, genesis, false)
// }

func initGenesisDposContext(g *Genesis, db badger.IStorage) *avlmerkle.DposContext {
	dc, err := avlmerkle.NewDposContextFromProto(db, &avlmerkle.DposContextProto{})
	if err != nil {
		return nil
	}
	// dc.SetValidators
	fmt.Print(len(g.Config.Dpos.Validators))
	// fmt.Printf("%v %v %v\n", g.Config, g.Config.Dpos, g.Config.Dpos.Validators)
	if len(g.Config.Dpos.Validators) > 0 {
		dc.SetValidators(g.Config.Dpos.Validators)
		for _, validator := range g.Config.Dpos.Validators {
			dc.DelegateTrie().Update(append(validator.Bytes(), validator.Bytes()...), validator.Bytes())
			dc.CandidateTrie().Update(validator.Bytes(), validator.Bytes())
		}
	}
	return dc
}

func NewGenesis(config *GenesisConfig, chainConfig *params.ChainConfig, bits uint32) *Genesis {
	account := make(map[string]string)
	account[config.GenesisCoinbase] = config.Balance
	return &Genesis{
		Config:        chainConfig,
		GenesisConfig: *config,
		Nonce:         0,
		Coinbase:      config.GenesisCoinbase,
		Accounts:      account,
		Bits:          bits,
	}
}

func (g *Genesis) WriteGenesisBlockN() (*Block, error) {

	// Genesis specifies the header fields, state of a genesis block. It also defines accounts
	chaindb := newChainDBN(g.ChainDB, g.Debug)
	stateTree := NewStateTree(g.StateDB, nil)
	//logrus.Debugf("initialize genesis account count: %d", len(genesis.Accounts))
	for addr, a := range g.Accounts {
		address := common.B58ToAddress([]byte(addr))
		balance := common.ParseString2BigInt(a)
		stateTree.AddBalance(address, balance)
		//logrus.Debugf("initialize genesis account: %s, balance: %d", address, balance)
	}
	stateTree.UpdateAll()
	timestamp := common.ParseString2BigInt(g.Timestamp)
	var coinbase common.Address
	if g.Coinbase != "" {
		coinbase = common.B58ToAddress([]byte(g.Coinbase))
	}
	GenesisBits = g.Bits
	rootHash := common.Bytes2Hash(stateTree.Root())
	HashPrevBlock := common.Hex2Hash(g.HashPrevBlock)

	// add dposcontext

	dposContext := initGenesisDposContext(g, g.ChainDB)
	dposContextProto := dposContext.ToProto()
	block := NewBlock(&BlockHeader{
		Nonce:         g.Nonce,
		ExtraNonce:    g.ExtraNonce,
		HashPrevBlock: HashPrevBlock,
		Timestamp:     timestamp.Uint64(),
		Coinbase:      coinbase,
		GasUsed:       new(big.Int),
		GasLimit:      common.GenesisGasLimit,
		Bits:          g.Bits,
		StateRoot:     rootHash,
		DposContext:   dposContextProto,
	}, nil, nil)

	if old := chaindb.GetBlockHeaderByHash(block.HeaderHash()); old != nil {
		logrus.WithField("hash", old.HashHex()).Infof("Genesis Block")
		oldGeneisBlock := &Block{Header: old, Transactions: nil, Receipts: nil}
		return oldGeneisBlock, nil
	}
	logrus.WithField("hash", block.HashHex()).Infof("Write genesis block")
	//logrus.Infof("write genesis block hash: %s", block.HashHex())
	if err := stateTree.Commit(); err != nil {
		return nil, err
	}
	if err := chaindb.WriteBHeaderWithHash(block.Header); err != nil {
		return nil, err
	}
	if err := chaindb.WriteBHeader2Chain(block.Header); err != nil {
		return nil, err
	}
	block.DposContext = dposContext
	return block, nil
}

func (g *Genesis) WriteMainNetGenesisBlockN() (*Block, error) {
	g.Coinbase, g.Balance = "haF8HrbHByusg6VcCqdjZqMrKasKNv7KN", "0"
	g.Bits = MainNetGenesisBits
	return g.WriteGenesisBlockN()
}

func (g *Genesis) WriteTestNetGenesisBlockN() (*Block, error) {
	g.Coinbase, g.Balance = "bQfi7kVUf2VAUsBk1R9FEzHXdtNtD98bs", "600000000000000000000000000"
	g.Bits = TestNetGenesisBits
	return g.WriteGenesisBlockN()
}
