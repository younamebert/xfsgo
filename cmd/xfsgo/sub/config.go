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

package sub

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"xfsgo"
	"xfsgo/backend"
	"xfsgo/common"
	"xfsgo/node"

	"github.com/spf13/viper"
)

const (
	defaultConfigFile        = "./config.yml"
	defaultStorageDir        = ".xfsgo"
	defaultChainDir          = "chain"
	defaultStateDir          = "state"
	defaultKeysDir           = "keys"
	defaultExtraDir          = "extra"
	defaultNodesDir          = "nodes"
	defaultRPCClientAPIHost  = "127.0.0.1:9012"
	defaultNodeRPCListenAddr = "127.0.0.1:9012"
	defaultNodeP2PListenAddr = "0.0.0.0:9011"
	defaultNetworkId         = uint32(1)
	defaultTestNetworkId     = uint32(2)
	defaultProtocolVersion   = uint32(1)
	defaultLoggerLevel       = "INFO"
	defaultCliTimeOut        = "180s"
	defaultNodeSyncFlag      = true
)

var defaultMinGasPrice = common.DefaultGasPrice()

// var defaultMaxGasLimit = common.MinGasLimit

type storageParams struct {
	dataDir  string
	chainDir string
	keysDir  string
	stateDir string
	extraDir string
	nodesDir string
}

type loggerParams struct {
	level string
}

type daemonConfig struct {
	loggerParams  loggerParams
	storageParams storageParams
	nodeConfig    node.Config
	backendParams backend.Params
}

type clientConfig struct {
	rpcClientApiHost    string
	rpcClientApiTimeOut string
}

var defaultNumWorkers = uint32(runtime.NumCPU())

func readFromConfigPath(v *viper.Viper, customFile string) error {
	filename := filepath.Base(defaultConfigFile)
	ext := filepath.Ext(defaultConfigFile)
	configPath := filepath.Dir(defaultConfigFile)
	v.AddConfigPath("$HOME/.xfsgo")
	v.AddConfigPath("/etc/xfsgo")
	v.AddConfigPath(configPath)
	v.SetConfigType(strings.TrimPrefix(ext, "."))
	v.SetConfigName(strings.TrimSuffix(filename, ext))
	v.SetConfigFile(customFile)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func parseConfigLoggerParams(v *viper.Viper) loggerParams {
	params := loggerParams{}
	params.level = v.GetString("logger.level")
	if params.level == "" {
		params.level = defaultLoggerLevel
	}
	return params
}

func setupDataDir(params *storageParams, datadir string) {
	if datadir != "" && params.dataDir != datadir {
		np := new(storageParams)
		np.dataDir = datadir
		*params = *np
	}
	if params.chainDir == "" {
		params.chainDir = filepath.Join(
			params.dataDir, defaultChainDir)
	}
	if params.stateDir == "" {
		params.stateDir = filepath.Join(
			params.dataDir, defaultStateDir)
	}
	if params.keysDir == "" {
		params.keysDir = filepath.Join(
			params.dataDir, defaultKeysDir)
	}
	if params.extraDir == "" {
		params.extraDir = filepath.Join(
			params.dataDir, defaultExtraDir)
	}
	if params.nodesDir == "" {
		params.nodesDir = filepath.Join(
			params.dataDir, defaultNodesDir)
	}
}

func parseConfigStorageParams(v *viper.Viper) storageParams {
	params := storageParams{}
	params.dataDir = v.GetString("storage.datadir")
	params.chainDir = v.GetString("storage.chaindir")
	params.stateDir = v.GetString("storage.statedir")
	params.keysDir = v.GetString("storage.keysdir")
	params.extraDir = v.GetString("storage.extradir")
	params.nodesDir = v.GetString("storage.nodesdir")
	if params.dataDir == "" {
		home := os.Getenv("HOME")
		params.dataDir = filepath.Join(
			home, defaultStorageDir)
	}
	setupDataDir(&params, params.dataDir)
	return params
}

func getDefaultNodeSyncFlag() bool {
	return defaultNodeSyncFlag
}

func defaultBootstrapNodes(netid uint32) []string {
	// hardcoded bootstrap nodes
	if netid == 1 {
		// main net boot nodes
		return []string{}
	}
	if netid == 2 {
		// test net boot nodes
		return []string{
			"xfsnode://45.63.126.195:9011/?id=66fe5febed75340c88ab00df0c7760e1682b48471d25b3e160eb140cb173422cc4d6dba2921e07d5826d880ede01ecbea32ec4dcc7fb5e7ff16cc34967583319",
			"xfsnode://119.28.26.67:9011/?id=61d6c98ad5c63db081734667d20dc8de4c6050d2e1ca01b28caa55d0717e5dac07b2b07fe1bcdd15bcf223e2c7146362713baaff5afb96f405d7734ef3977b68",
			"xfsnode://78.141.192.47:9011/?id=aeb058ca7936d66a59834205e2ef559b14aa37c14578abd8ecc8e64d2d1e58d134dee828814ca1a52101285caa94eb949b185cdd6d8d936e8f41a0beca00845f",
			"xfsnode://45.32.88.212:9011/?id=031fb41343d89c27ffe48bc834c39e3f6d9452435da40091893f5594b674104cb0e9e76bb6f2ec73e28bcfebd5ae80fa46622d81b2bc092e1800cc5f42ea72e1",
			"xfsnode://45.32.243.66:9011/?id=d7ad667628482d1dc611860b2b952f6433c90a2ae110f62b96ed439f13c6ae6c8c9ff1ab2c2e5972c75c6849b4d2ce085ce837e4d5e4da77441abf4ece5a28d6",
		}
	}
	return make([]string, 0)
}
func parseConfigNodeParams(v *viper.Viper, netid uint32) node.Config {
	config := node.Config{
		RPCConfig: new(xfsgo.RPCConfig),
	}
	config.RPCConfig.ListenAddr = v.GetString("rpcserver.listen")
	config.P2PListenAddress = v.GetString("p2pnode.listen")
	config.P2PBootstraps = v.GetStringSlice("p2pnode.bootstrap")
	config.P2PStaticNodes = v.GetStringSlice("p2pnode.static")
	config.ProtocolVersion = uint8(v.GetUint64("protocol.version"))
	config.NodeSyncFlag = v.GetBool("p2pnode.syncflag")

	if config.RPCConfig.ListenAddr == "" {
		config.RPCConfig.ListenAddr = defaultNodeRPCListenAddr
	}
	if config.P2PListenAddress == "" {
		config.P2PListenAddress = defaultNodeP2PListenAddr
	}
	if config.P2PBootstraps == nil || len(config.P2PBootstraps) == 0 {
		config.P2PBootstraps = defaultBootstrapNodes(netid)
	}

	if config.NodeSyncFlag || getDefaultNodeSyncFlag() {
		config.NodeSyncFlag = true
	}

	return config
}

func parseConfigBackendParams(v *viper.Viper) backend.Params {
	config := backend.Params{}
	mCoinbase := v.GetString("miner.coinbase")
	if mCoinbase != "" {
		config.Coinbase = common.StrB58ToAddress(mCoinbase)
	}
	minGasPriceStr := v.GetString("miner.gasprice")
	config.ProtocolVersion = v.GetUint32("protocol.version")
	config.NetworkID = v.GetUint32("protocol.networkid")
	config.Numworkers = v.GetUint32("miner.numworkers")
	var minGasPrice *big.Int
	var ok bool
	if minGasPrice, ok = new(big.Int).SetString(minGasPriceStr, 10); !ok || minGasPrice.Cmp(defaultMinGasPrice) < 0 {
		minGasPrice = defaultMinGasPrice
	}
	config.MinGasPrice = minGasPrice
	if config.Numworkers == uint32(0) {
		config.Numworkers = defaultNumWorkers
	}

	if config.ProtocolVersion == 0 {
		config.ProtocolVersion = defaultProtocolVersion
	}
	if config.NetworkID == 0 {
		config.NetworkID = defaultNetworkId
	}
	config.GenesisFile = v.GetString("protocol.genesisfile")
	return config
}

func parseDaemonConfig(configFilePath string) (daemonConfig, error) {
	config := viper.New()
	if err := readFromConfigPath(config, configFilePath); err != nil {
	}
	mStorageParams := parseConfigStorageParams(config)
	mBackendParams := parseConfigBackendParams(config)
	mLoggerParams := parseConfigLoggerParams(config)
	nodeParams := parseConfigNodeParams(config, mBackendParams.NetworkID)
	nodeParams.NodeDBPath = mStorageParams.nodesDir
	return daemonConfig{
		loggerParams:  mLoggerParams,
		storageParams: mStorageParams,
		nodeConfig:    nodeParams,
		backendParams: mBackendParams,
	}, nil
}

func parseClientConfig(configFilePath string) (clientConfig, error) {
	config := viper.New()
	if err := readFromConfigPath(config, configFilePath); err != nil {
	}
	mRpcClientApiHost := config.GetString("rpclient.apihost")
	if rpchost == "" {
		if mRpcClientApiHost == "" {
			mRpcClientApiHost = fmt.Sprintf("http://%s", defaultRPCClientAPIHost)
		}
	} else {
		mRpcClientApiHost = fmt.Sprintf("http://%s", rpchost)
	}
	mRpcClientApiTimeOut := config.GetString("rpclient.timeout")
	if mRpcClientApiTimeOut == "" {
		mRpcClientApiTimeOut = defaultCliTimeOut
	}
	timeDur, err := time.ParseDuration(mRpcClientApiTimeOut)
	if err != nil {
		return clientConfig{}, err
	}
	times := timeDur.Seconds()
	if times < 1 && times > 3*60 {
		return clientConfig{}, fmt.Errorf("time overflow")
	}
	return clientConfig{
		rpcClientApiHost:    mRpcClientApiHost,
		rpcClientApiTimeOut: mRpcClientApiTimeOut,
	}, nil
}
