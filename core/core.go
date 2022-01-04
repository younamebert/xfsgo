package core

import (
	"xfsgo"
	"xfsgo/common"
	"xfsgo/consensus"
	"xfsgo/consensus/dpos"
)

type CoreChain struct {
	Chain  xfsgo.IBlockChain
	Engine consensus.Engine
}

// InsertChain executes the actual chain insertion.
func (c *CoreChain) InsertChain(block *xfsgo.Block) error {
	if err := c.Engine.VerifyHeader(c.Chain, block.GetHeader(), true); err != nil {
		return err
	}

	// Validate validator
	dposEngine, isDpos := c.Engine.(*dpos.Dpos)
	if isDpos {
		err := dposEngine.VerifySeal(c.Chain, block.GetHeader())
		if err != nil {
			return err
		}
	}
	if err := c.Chain.InsertChain(block); err != nil {
		return err
	}
	return nil
}

func (c *CoreChain) CurrentBHeader() *xfsgo.BlockHeader {
	return c.Chain.CurrentBHeader()
}

func (c *CoreChain) GenesisBHeader() *xfsgo.BlockHeader {
	return c.Chain.GenesisBHeader()
}

func (c *CoreChain) GetBlockByNumber(num uint64) *xfsgo.Block {
	return c.Chain.GetBlockByNumber(num)
}

func (c *CoreChain) GetBlockHashesFromHash(hash common.Hash, max uint64) []common.Hash {
	return c.Chain.GetBlockHashesFromHash(hash, max)
}

func (c *CoreChain) GetBlockByHashWithoutRec(hash common.Hash) *xfsgo.Block {
	return c.Chain.GetBlockByHashWithoutRec(hash)
}

func (c *CoreChain) GetReceiptByHash(hash common.Hash) *xfsgo.Receipt {
	return c.Chain.GetReceiptByHash(hash)
}

func (c *CoreChain) GetBlockByHash(hash common.Hash) *xfsgo.Block {
	return c.Chain.GetBlockByHash(hash)
}

func (c *CoreChain) SetBoundaries(syncStatsOrigin, syncStatsHeight uint64) error {
	return c.Chain.SetBoundaries(syncStatsOrigin, syncStatsHeight)
}
