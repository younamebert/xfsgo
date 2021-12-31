package core

import (
	"xfsgo"
	"xfsgo/consensus"
)

type Engine struct {
	blockchain xfsgo.IBlockChain
	engine     consensus.Engine
}
