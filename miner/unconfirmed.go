package miner

import (
	"container/ring"
	"sync"
	"xfsgo"
	"xfsgo/common"

	"github.com/sirupsen/logrus"
)

// headerRetriever is used by the unconfirmed block set to verify whether a previously
// mined block is part of the canonical chain or not.
// type headerRetriever interface {
// 	// GetHeaderByNumber retrieves the canonical header associated with a block number.
// 	GetHeaderByNumber(number uint64) *xfsgo.BlockHeader
// }

// unconfirmedBlock is a small collection of metadata about a locally mined block
// that is placed into a unconfirmed set for canonical chain inclusion tracking.
type unconfirmedBlock struct {
	index uint64
	hash  common.Hash
}

// unconfirmedBlocks implements a data structure to maintain locally mined blocks
// have have not yet reached enough maturity to guarantee chain inclusion. It is
// used by the miner to provide logs to the user when a previously mined block
// has a high enough guarantee to not be reorged out of te canonical chain.
type unconfirmedBlocks struct {
	chain  xfsgo.IBlockChain // Blockchain to verify canonical status through
	depth  uint              // Depth after which to discard previous blocks
	blocks *ring.Ring        // Block infos to allow canonical chain cross checks
	lock   sync.RWMutex      // Protects the fields from concurrent access
}

// newUnconfirmedBlocks returns new data structure to track currently unconfirmed blocks.
func newUnconfirmedBlocks(chain xfsgo.IBlockChain, depth uint) *unconfirmedBlocks {
	return &unconfirmedBlocks{
		chain: chain,
		depth: depth,
	}
}

// Insert adds a new block to the set of unconfirmed ones.
func (set *unconfirmedBlocks) Insert(index uint64, hash common.Hash) {
	// If a new block was mined locally, shift out any old enough blocks
	set.Shift(index)

	// Create the new item as its own ring
	item := ring.New(1)
	item.Value = &unconfirmedBlock{
		index: index,
		hash:  hash,
	}
	// Set as the initial ring or append to the end
	set.lock.Lock()
	defer set.lock.Unlock()

	if set.blocks == nil {
		set.blocks = item
	} else {
		set.blocks.Move(-1).Link(item)
	}
	// Display a log for the user to notify of a new mined block unconfirmed
	logrus.Infof(" mined potential block number:%v hash:%v\n", index, hash.Hex())
}

// Shift drops all unconfirmed blocks from the set which exceed the unconfirmed sets depth
// allowance, checking them against the canonical chain for inclusion or staleness
// report.
func (set *unconfirmedBlocks) Shift(height uint64) {
	set.lock.Lock()
	defer set.lock.Unlock()

	for set.blocks != nil {
		// Retrieve the next unconfirmed block and abort if too fresh
		next := set.blocks.Value.(*unconfirmedBlock)
		if next.index+uint64(set.depth) > height {
			break
		}
		// Block seems to exceed depth allowance, check for canonical status
		header := set.chain.GetBlockByNumber(next.index).Header
		switch {
		case header == nil:
			logrus.Warnf("Failed to retrieve header of mined block number %s hash %s", next.index, next.hash.Hex())
		case header.HashHex() == next.hash.Hex():
			logrus.Infof("block reached canonical chain number %s hash %s", next.index, next.hash.Hex())
		default:
			logrus.Infof("⑂ block  became a side fork number %s hash %s", next.index, next.hash.Hex())
		}
		// Drop the block out of the ring
		if set.blocks.Value == set.blocks.Next().Value {
			set.blocks = nil
		} else {
			set.blocks = set.blocks.Move(-1)
			set.blocks.Unlink(1)
			set.blocks = set.blocks.Move(1)
		}
	}
}
