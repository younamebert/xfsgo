package miner

import (
	"testing"
	"time"
	"xfsgo"
	"xfsgo/consensus/dpos"
	"xfsgo/test"
)

//func TestMinerStartStop(t *testing.T) {
//	miner := createMiner(t)
//	waitForMiningState(t, miner, false)
//	miner.Start(test.TestMinerWorkers)
//	waitForMiningState(t, miner, true)
//	miner.Stop()
//	waitForMiningState(t, miner, false)
//}
//
//func TestSetWorkers(t *testing.T) {
//	miner := createMiner(t)
//	waitForMiningState(t, miner, false)
//	miner.Start(test.TestMinerWorkers)
//	miner.SetWorkers(uint32(5))
//	waitForMiningState(t, miner, true)
//	miner.Stop()
//	waitForMiningState(t, miner, false)
//}

func waitForMiningState(t *testing.T, m *Miner, mining bool) {
	t.Helper()
	var state bool
	for i := 0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		if state = m.GetMinStatus(); state == mining {
			return
		}
	}
	t.Fatalf("Mining() == %t, want %t", state, mining)
}

func createMiner(t *testing.T) *Miner {

	stateDb := test.NewMemStorage()

	chainDb := test.NewMemStorage()

	extraDb := test.NewMemStorage()

	keyDb := test.NewMemStorage()

	wallet := xfsgo.NewWallet(keyDb)

	event := xfsgo.NewEventBus()

	testChainConfig := test.TestChainConfig

	initConf := &xfsgo.GenesisConfig{
		StateDB: stateDb,
		ChainDB: chainDb,
		Debug:   false,
	}
	testGenesis := xfsgo.NewGenesis(initConf, testChainConfig, test.TestGenesisBits)
	if _, err := testGenesis.WriteTestGenesisBlockN(); err != nil {
		t.Error(err)
		return nil
	}
	bc, err := xfsgo.NewBlockChainN(stateDb, chainDb, extraDb, event, testChainConfig, false)
	if err != nil {
		t.Error(err)
		return nil
	}

	testCoinbase, err := wallet.AddByRandom()
	if err != nil {
		t.Errorf("new wallet account err:%v", err)
		return nil
	}
	testChainConfig.Dpos.Validators = append(testChainConfig.Dpos.Validators, testCoinbase)

	txPool := xfsgo.NewTxPool(bc.CurrentStateTree, bc.LatestGasLimit, test.TestTxPoolGasPrice, event)
	config := &Config{
		Coinbase:   testCoinbase,
		Numworkers: test.TestMinerWorkers,
	}

	testdpos := dpos.New(testChainConfig.Dpos, chainDb)

	return NewMiner(config, wallet.All(), stateDb, bc, event, txPool, test.TestTxPoolGasPrice, test.TestTxPoolGasLimit, testdpos, chainDb)

}
