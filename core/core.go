package core

import (
	"xfsgo"
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
