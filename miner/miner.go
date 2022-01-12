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

package miner

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
	"xfsgo"
	"xfsgo/avlmerkle"
	"xfsgo/common"
	"xfsgo/consensus"
	"xfsgo/consensus/dpos"
	"xfsgo/crypto"
	"xfsgo/storage/badger"

	"github.com/sirupsen/logrus"
)

// const (
// 	// hpsUpdateSecs is the number of seconds to wait in between each
// 	// update to the hashes per second monitor.
// 	hashUpdateSecs = 1

// 	// hashUpdateSec is the number of seconds each worker waits in between
// 	// notifying the speed monitor with how many hashes have been completed
// 	// while they are actively searching for a solution.  This is done to
// 	// reduce the amount of syncs between the workers that must be done to
// 	// keep track of the hashes per second.
// 	hpsUpdateSecs = 10

// 	// maxNonce is the maximum value a nonce can be in a block header.
// 	maxNonce = ^uint32(0) // 2^32 - 1

// 	// maxExtraNonce is the maximum value an extra nonce used in a coinbase
// 	// transaction can be.
// 	maxExtraNonce = ^uint64(0) // 2^64 - 1
// )

// var (
// 	hashRateLoopIntervalSec = 1
// 	progressReportTime      = 10 * time.Second
// 	maxWorkers              = uint32(255)
// 	defaultNumWorkers       = uint32(runtime.NumCPU())
//  applyTransactionsErr = errors.New("apply transaction err")

// )

const (
	miningLogAtDepth = 5
	resultQueueSize  = 10
)

type Config struct {
	Coinbase   common.Address
	Validator  common.Address
	Numworkers uint32
}

type Result struct {
	Work  *Work
	Block *xfsgo.Block
}

type Work struct {
	Chain xfsgo.IBlockChain

	state       *xfsgo.StateTree // apply state changes here
	dposContext *avlmerkle.DposContext

	Block *xfsgo.Block // the new block

	header    *xfsgo.BlockHeader
	txs       []*xfsgo.Transaction
	receipts  []*xfsgo.Receipt
	createdAt time.Time
}

// Miner creates blocks with transactions in tx pool and searches for proof-of-work values.
type Miner struct {
	*Config
	mu      sync.Mutex
	rwmu    sync.RWMutex
	started bool
	quitCh  chan struct{}
	sealCh  chan struct{}
	// runningWorkers   []chan struct{}
	// updateNumWorkers chan uint32
	numWorkers    uint32
	eventBus      *xfsgo.EventBus
	canStart      bool
	shouldStart   bool
	pool          *xfsgo.TxPool
	chain         xfsgo.IBlockChain
	stateDb       badger.IStorage
	accounts      map[common.Address]*ecdsa.PrivateKey
	gasPrice      *big.Int
	gasLimit      *big.Int
	LastStartTime time.Time
	rmlock        sync.RWMutex
	remove        map[common.Hash]*xfsgo.Transaction
	wg            sync.WaitGroup
	// workerWg        sync.WaitGroup
	// runningHashRate chan common.HashRate
	// lastHashRate    common.HashRate
	// reportHashes    chan uint64

	engine      consensus.Engine
	chainDb     badger.IStorage
	recv        chan *Result
	current     *Work
	unconfirmed *unconfirmedBlocks // set of locally mined blocks pending canonicalness confirmations
}

func NewMiner(config *Config,
	accounts map[common.Address]*ecdsa.PrivateKey,
	stateDb *badger.Storage,
	chain xfsgo.IBlockChain,
	eventBus *xfsgo.EventBus,
	pool *xfsgo.TxPool,
	gasPrice, gasLimit *big.Int,
	engine consensus.Engine,
	chainDb badger.IStorage) *Miner {

	m := &Miner{
		Config:     config,
		chain:      chain,
		stateDb:    stateDb,
		accounts:   make(map[common.Address]*ecdsa.PrivateKey),
		numWorkers: 0,
		// updateNumWorkers: make(chan uint32),
		pool:        pool,
		canStart:    true,
		shouldStart: false,
		started:     false,
		eventBus:    eventBus,
		gasLimit:    gasLimit,
		gasPrice:    gasPrice,
		remove:      make(map[common.Hash]*xfsgo.Transaction),
		// reportHashes:    make(chan uint64, 1),
		// runningHashRate: make(chan common.HashRate),
		unconfirmed: newUnconfirmedBlocks(chain, miningLogAtDepth),
		engine:      engine,
		chainDb:     chainDb,
		recv:        make(chan *Result, resultQueueSize),
		quitCh:      make(chan struct{}, 1),
		sealCh:      make(chan struct{}, 1),
	}
	m.accounts = accounts
	m.LoadLauncher()
	go m.update()
	return m
}

func (m *Miner) isValidator() (common.Address, error) {
	m.rwmu.RLock()
	validator := m.Validator
	m.rwmu.RUnlock()
	if validator != (common.Address{}) {
		return validator, nil
	}
	for addr := range m.accounts {
		return addr, nil
	}
	return common.Address{}, fmt.Errorf("validator address must be explicitly specified")
}

func (m *Miner) SetValidator(addr common.Address) bool {
	m.Validator = addr
	return true
}

func (m *Miner) HashRate() string {
	return "0"
}

func (m *Miner) update() {
	startSub := m.eventBus.Subscript(xfsgo.SyncStartEvent{})
	doneSub := m.eventBus.Subscript(xfsgo.SyncDoneEvent{})
	failedSub := m.eventBus.Subscript(xfsgo.SyncFailedEvent{})
	defer func() {
		startSub.Unsubscribe()
		doneSub.Unsubscribe()
		failedSub.Unsubscribe()
	}()
out:
	for {
		select {
		case <-startSub.Chan():
			m.mu.Lock()
			m.canStart = false
			m.mu.Unlock()
			m.Stop()
		case <-failedSub.Chan():
			m.mu.Lock()
			m.canStart = true
			m.mu.Unlock()
			// if m.shouldStart {
			// 	m.Start(m.numWorkers)
			// }
		case <-doneSub.Chan():
			m.mu.Lock()
			m.canStart = true
			m.mu.Unlock()
			// if m.shouldStart {
			// 	m.Start(m.numWorkers)
			// }
			break out
		}
	}
}

func (m *Miner) GetWorkerNum() uint32 {
	return m.numWorkers
}

func (m *Miner) GetGasPrice() *big.Int {
	return m.gasPrice
}

func (m *Miner) GetGasLimit() *big.Int {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()
	return m.gasLimit
}

// func (m *Miner) GetWorkerNum() uint32 {
// 	return m.numWorkers
// }

func (m *Miner) GetMinStatus() bool {
	return m.started
}

func (m *Miner) GetNext() bool {
	return m.started
}

// func (m *Miner) RunningHashRate() common.HashRate {
// 	select {
// 	case r := <-m.runningHashRate:
// 		return r
// 	default:
// 	}
// 	return m.lastHashRate
// }

func (m *Miner) SetGasLimit(limit *big.Int) error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()
	if limit.Cmp(common.MinGasLimit) < 0 {
		return errors.New("gas limit too low")
	} else if limit.Cmp(common.GenesisGasLimit) > 0 {
		return errors.New("gas limit out of GenesisGasLimit")
	} else if limit.Cmp(m.gasLimit) == 0 {
		return nil
	}
	m.gasLimit = limit
	return nil
}

func (m *Miner) SetGasPrice(price *big.Int) error {
	if price.Cmp(common.Big1) < 0 {
		return errors.New("gas price too low")
	} else if price.Cmp(m.gasPrice) == 0 {
		return nil
	}
	gasPriceChanged := &xfsgo.GasPriceChanged{
		Price: price,
	}
	m.gasPrice = price
	m.eventBus.Publish(gasPriceChanged)
	return nil
}

func (m *Miner) mintLoop() {
	ticker := time.NewTicker(time.Second).C
out:
	for {
		select {
		case now := <-ticker:
			m.mintBlock(now.Unix())
		case <-m.quitCh:
			close(m.sealCh)
			m.started = false
			m.quitCh = make(chan struct{}, 1)
			m.sealCh = make(chan struct{}, 1)
			break out
		}
	}
	m.wg.Done()
	logrus.Info("Miner quit")
}

func (m *Miner) waitworker() {
	for {
		for result := range m.recv {
			// atomic.AddInt32(&self.atWork, -1)

			if result == nil || result.Block == nil {
				continue
			}
			block := result.Block

			// work := result.Work

			// addr := block.Coinbase()
			// keyhash := ahash.SHA256(addr.Bytes())
			// m.stateDb.Foreach(func(k string, v []byte) error {
			// 	fmt.Printf("key:%v v:%v\n", k, string(v))
			// 	return nil
			// })

			if err := m.chain.WriteBlock(block); err != nil {
				logrus.Errorf("Failed writing block to chain err:%v", err)
				continue
			}

			if _, err := block.DposContext.CommitTo(); err != nil {
				logrus.Errorf("Failed writing block to chain DposContext err:%v", err)
				continue
			}

			bcHash := block.HeaderHash()
			m.eventBus.Publish(xfsgo.NewMinedBlockEvent{Block: block})
			m.eventBus.Publish(xfsgo.NewBlockEvent{Block: block})
			// Insert the block into the set of pending ones to wait for confirmations
			m.unconfirmed.Insert(block.Header.Number().Uint64(), bcHash)
			logrus.Infof("Successfully sealed new block number:%s hash:%s\n", block.Header.Number(), bcHash.Hex())
		}
	}
}

func (m *Miner) updateWorker() {
	txPreEventSub := m.eventBus.Subscript(xfsgo.TxPreEvent{})
	newBlockEventSub := m.eventBus.Subscript(xfsgo.NewBlockEvent{})
	defer newBlockEventSub.Unsubscribe()
	defer txPreEventSub.Unsubscribe()

	for {
		select {
		case <-newBlockEventSub.Chan():
			close(m.sealCh)
			m.sealCh = make(chan struct{})
		case e := <-txPreEventSub.Chan():
			event, ok := e.(xfsgo.TxPreEvent)
			if !ok {
				return
			}
			Tx := event.Tx
			_ = m.pool.Add(Tx)

			txs := m.pool.GetTransactions()

			lastBlock := m.chain.CurrentBHeader()
			lastStateRoot := lastBlock.StateRoot
			//lastBlockHash := lastBlock.Hash()
			//logrus.Debugf("Generating block by parent height=%d, hash=0x%x...%x, workerId=%-3d", lastBlock.Height(), lastBlockHash[:4], lastBlockHash[len(lastBlockHash)-4:], num)
			stateTree := xfsgo.NewStateTree(m.stateDb, lastStateRoot.Bytes())

			committx := make([]*xfsgo.Transaction, 0)
			ignoretxs := make(map[common.Address]struct{})
			gasused, res, err := m.applyTransactions(
				m.current.state, m.current.header, txs, ignoretxs, &committx)
			if err != nil {
				m.doRemove()
			}

			cugasUsed := m.current.header.GasUsed
			m.current.header.GasUsed = gasused.Add(gasused, cugasUsed)

			dpos.AccumulateRewards(m.chain.Config(), m.current.state, m.current.header)

			stateTree.UpdateAll()
			m.current.txs = append(m.current.txs, committx...)
			m.current.receipts = append(m.current.receipts, res...)
			m.current.Block.Transactions = m.current.txs
			m.current.Block.Receipts = m.current.receipts

		}
	}
}

func (m *Miner) mintBlock(now int64) {

	engine, ok := m.engine.(*dpos.Dpos)
	if !ok {
		logrus.Error("Only the dpos engine was allowed")
		return
	}
	err := engine.CheckValidator(m.chain.CurrentBlock(), now)

	if err != nil {
		switch err {
		case dpos.ErrWaitForPrevBlock,
			dpos.ErrMintFutureBlock,
			dpos.ErrInvalidBlockValidator,
			dpos.ErrInvalidMintBlockTime:
			logrus.Debugf("Failed to mint the block, while err: %v\n", err)
		default:
			logrus.Errorf("Failed to mint the block err: %v\n", err)
		}
		return
	}
	work, err := m.createNewWork()
	if err != nil {
		logrus.Errorf("Failed to create the new work err:%v\n", err)
		return
	}

	result, err := m.engine.Seal(m.chain, work.Block, m.sealCh)
	if err != nil {
		logrus.Errorf("Failed to seal the block err:%v\n", err)
		return
	}
	m.recv <- &Result{work, result}
}

func (m *Miner) createNewWork() (*Work, error) {
	tstart := time.Now()
	parent := m.chain.CurrentBlock()
	tstamp := tstart.Unix()
	if parent.Time().Cmp(new(big.Int).SetInt64(tstamp)) >= 0 {
		tstamp = parent.Time().Int64() + 1
	}
	// m.stateDb.GetData()
	// this will ensure we're not going off too far in the future
	if now := time.Now().Unix(); tstamp > now+1 {
		wait := time.Duration(tstamp-now) * time.Second
		logrus.Info("Mining too far in the future", "wait", common.PrettyDuration(wait))
		time.Sleep(wait)
	}
	num := parent.Height()
	header := &xfsgo.BlockHeader{
		Height:        num + 1,
		GasLimit:      common.TxPoolGasLimit,
		GasUsed:       new(big.Int),
		HashPrevBlock: parent.HeaderHash(),
		Coinbase:      m.Coinbase,
		Timestamp:     uint64(tstamp),
	}
	header.GasUsed = new(big.Int)

	if err := m.engine.Prepare(m.chain, header); err != nil {
		return nil, fmt.Errorf("got error when preparing header, err: %s", err)
	}

	// Could potentially happen if starting to mine in an odd state.
	err := m.makeCurrent(parent, header)
	if err != nil {
		return nil, fmt.Errorf("got error when create mining context, err: %s", err)
	}

	work := m.current

	pending := m.pool.GetTransactions()

	xfsgo.SortByPriceAndNonce(pending)

	lastBlock := m.chain.CurrentBHeader()
	lastStateRoot := lastBlock.StateRoot
	//lastBlockHash := lastBlock.Hash()
	//logrus.Debugf("Generating block by parent height=%d, hash=0x%x...%x, workerId=%-3d", lastBlock.Height(), lastBlockHash[:4], lastBlockHash[len(lastBlockHash)-4:], num)
	stateTree := xfsgo.NewStateTree(m.stateDb, lastStateRoot.Bytes())

	logrus.Debugf("minerbal:%v coinbase:%v\n", stateTree.GetBalance(header.Coinbase).String(), header.Coinbase.B58String())

	committx := make([]*xfsgo.Transaction, 0)
	ignoretxs := make(map[common.Address]struct{})
	gasused, res, err := m.applyTransactions(
		stateTree, header, pending, ignoretxs, &committx)
	if err != nil {
		return nil, fmt.Errorf("apply transaction err:%v", err)
	}

	header.GasUsed.Set(gasused)

	// Create the new block to seal with the consensus engine
	block, err := m.engine.Finalize(m.chain, header, stateTree, committx, work.receipts, work.dposContext)

	if err != nil {
		return nil, fmt.Errorf("got error when finalize block for sealing, err: %s", err)
	}

	stateTree.UpdateAll()
	work.Block = block
	work.receipts = append(work.receipts, res...)
	work.Block.DposContext = work.dposContext

	logrus.Infof("Commit new mining work number %v  txs %v elapsed %v", work.Block.Height(), len(work.txs), common.PrettyDuration(time.Since(tstart)))
	m.unconfirmed.Shift(work.Block.Height())

	if err := stateTree.Commit(); err != nil {
		return nil, err
	}
	return work, nil
}

// makeCurrent creates a new environment for the current cycle.
func (m *Miner) makeCurrent(parent *xfsgo.Block, header *xfsgo.BlockHeader) error {
	state := m.chain.StateAt(parent.Header.Root())
	if state == nil {
		return errors.New("state root nil")
	}
	dposContext, err := avlmerkle.NewDposContextFromProto(m.chainDb, parent.Header.DposContext)
	if err != nil {
		return err
	}
	work := &Work{
		state:       state,
		header:      header,
		Block:       parent,
		dposContext: dposContext,
		txs:         parent.Transactions,
		receipts:    parent.Receipts,
		createdAt:   time.Now(),
	}

	m.current = work
	return nil
}

// Start starts up xfs mining
func (m *Miner) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started || !m.canStart {
		return
	}
	// m.quitCh = make(chan struct{})
	m.wg.Add(1)
	go m.mintLoop()
	m.LastStartTime = time.Now()
	m.started = true
	m.shouldStart = true
}

func (m *Miner) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		close(m.quitCh)
		m.wg.Wait()
	}
}

func (m *Miner) StartMining(threads *int) error {
	// Set the number of threads if the seal engine supports it
	if threads == nil {
		threads = new(int)
	} else if *threads == 0 {
		*threads = -1 // Disable the miner from within
	}
	m.numWorkers = uint32(*threads)
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := m.engine.(threaded); ok {
		logrus.Info("Updated mining threads", "threads", *threads)
		th.SetThreads(*threads)
	}

	validator, err := m.isValidator()
	if err != nil {
		logrus.Error("Cannot start mining without validator", "err", err)
		return fmt.Errorf("validator missing: %v", err)
	}
	if dpos, ok := m.engine.(*dpos.Dpos); ok {
		_, isaccount := m.accounts[validator]
		if !isaccount {
			return fmt.Errorf("coinbase account unavailable locally")
		}
		dpos.Authorize(validator, m.SignHash)
	}
	go m.Start()
	return nil
}

func (m *Miner) SignHash(account common.Address, hash []byte) ([]byte, error) {
	prikey, isaccount := m.accounts[account]
	if !isaccount {
		return nil, fmt.Errorf("coinbase account unavailable locally addr:%v", account.B58String())
	}
	return crypto.ECDSASign(hash, prikey)
}

// func (m *Miner) SetWorkers(num uint32) error {
// 	if num < 1 {
// 		return errors.New("number too low")
// 	} else if num > maxWorkers {
// 		return errors.New("number over max value")
// 	}
// 	m.updateNumWorkers <- num
// 	m.numWorkers = num
// 	return nil
// }

func (m *Miner) SetCoinbase(address common.Address) {
	m.Coinbase = address
}

// mainLoop is the miner's main event loop, waiting for and reacting to synchronize events.

func (m *Miner) appendRemove(tx *xfsgo.Transaction) {
	m.rmlock.Lock()
	defer m.rmlock.Unlock()
	txhash := tx.Hash()
	if _, exists := m.remove[txhash]; exists {
		return
	}
	m.remove[txhash] = tx
}
func (m *Miner) doRemove() {
	m.rmlock.RLock()
	defer m.rmlock.RUnlock()
	txs := make([]*xfsgo.Transaction, 0)
	for _, tx := range m.remove {
		txs = append(txs, tx)
		delete(m.remove, tx.Hash())
	}
	if len(txs) != 0 {
		m.pool.RemoveTransactions(txs)
	}
}

func (m *Miner) TargetHeight() uint64 {
	return m.chain.CurrentBHeader().Height + 1
}

func (m *Miner) NextDifficulty() float64 {
	bits, _ := m.chain.CalcNextRequiredDifficulty()
	return xfsgo.CalcDifficultyByBits(bits)
}

func (m *Miner) TargetHashRate() common.HashRate {
	bits, _ := m.chain.CalcNextRequiredDifficulty()
	return xfsgo.CalcHashRateByBits(bits)
}

func (m *Miner) applyTransactions(
	stateTree *xfsgo.StateTree,
	header *xfsgo.BlockHeader,
	txs []*xfsgo.Transaction,
	ignoreTxs map[common.Address]struct{},
	commitTxs *[]*xfsgo.Transaction) (*big.Int, []*xfsgo.Receipt, error) {
	receipts := make([]*xfsgo.Receipt, 0)
	var totalUsedGas uint64 = 0
	mGasPool := (*xfsgo.GasPool)(new(big.Int).Set(header.GasLimit))
	//pergp := (*big.Int)(mGasPool)
	//logrus.Debugf("Tx gas limit out of block limit-init: hash=%x, from=%x, mGasPool=%s", txfrom, pergp)

	snap := stateTree.Copy()
	dposSnap := m.current.dposContext.Snapshot()
	for _, tx := range txs {
		txfrom, _ := tx.FromAddr()
		txhash := tx.Hash()
		_ = txhash
		if _, exists := ignoreTxs[txfrom]; exists {
			//logrus.Warnf("Tx exists ignore obj: hash=%x, from=%x",
			//	txhash[len(txhash)-4:], txfrom)
			continue
		}
		rec, err := xfsgo.ApplyTransaction(m.chain, nil, mGasPool, stateTree, header, tx, &totalUsedGas, *m.chain.GetVMConfig(), m.current.dposContext)
		// rec, err := m.chain.ApplyTransaction(stateTree, header, tx, mGasPool, totalUsedGas)
		if err != nil {
			if err.Error() == xfsgo.ErrGasPoolOutErr.Error() {
				//logrus.Errorf("Miner apply transaction err will be ignore: %s", err)
				ignoreTxs[txfrom] = struct{}{}
				continue
			}
			logrus.Warnf("Miner  will be remove: %s", err)
			m.appendRemove(tx)
			stateTree.Set(snap)
			m.current.dposContext.RevertToSnapShot(dposSnap)
			return nil, nil, err
		}
		if rec != nil {
			receipts = append(receipts, rec)
		}
		//logrus.Debugf("Commit tx: %x",txhash[len(txhash)-4:])
		*commitTxs = append(*commitTxs, tx)
	}
	return new(big.Int).SetUint64(totalUsedGas), receipts, nil
}

func (m *Miner) LoadLauncher() {
	go m.updateWorker()
	go m.waitworker()
	m.createNewWork()
}
