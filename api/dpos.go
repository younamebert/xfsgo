package api

import (
	"math/big"
	"xfsgo"
	"xfsgo/avlmerkle"
	"xfsgo/common"
	"xfsgo/consensus/dpos"
	"xfsgo/storage/badger"
)

type DposAPIHandler struct {
	Chain   *xfsgo.BlockChain
	ChainDb *badger.Storage
	Dpos    *dpos.Dpos
}

type GetBlockNumByValidatorArgs struct {
	Number string `json:"number"`
}

// GetValidators retrieves the list of the validators at specified block
func (handler *DposAPIHandler) GetValidators(args GetBlockNumByValidatorArgs, resp *[]common.Address) error {
	var header xfsgo.IBlockHeader
	var last uint64
	if args.Number == "" {
		last = handler.Chain.CurrentBHeader(
	} else {
		number, ok := new(big.Int).SetString(args.Number, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
		last = number.Uint64()
	}

	header = handler.Chain.GetBlockByNumber(last).Header

	epochTrie, err := avlmerkle.NewEpochTrie(header.DposContext.EpochHash, handler.ChainDb)
	if err != nil {
		return err
	}
	dposContext := avlmerkle.DposContext{}
	dposContext.SetEpoch(epochTrie)
	validators, err := dposContext.GetValidators()
	if err != nil {
		return err
	}

	result := make([]common.Address, 0)
	for i := 0; i < len(validators); i++ {
		result = append(result, validators[i])
	}
	*resp = result
	return nil
}

// GetConfirmedBlockNumber retrieves the latest irreversible block
func (handler *DposAPIHandler) GetConfirmedBlockNumber(_ EmptyArgs, resp *int) error {
	var err error
	header := handler.Dpos.ConfirmedBlockHeader()
	if header == nil {
		header, err = handler.Dpos.LoadConfirmedBlockHeader(handler.Chain)
		if err != nil {
			return err
		}
	}
	*resp = int(header.Height)
	return nil
}
