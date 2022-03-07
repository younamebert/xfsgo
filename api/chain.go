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

package api

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strconv"
	"xfsgo"
	"xfsgo/common"
	"xfsgo/crypto"

	"github.com/sirupsen/logrus"
)

type ChainAPIHandler struct {
	BlockChain    *xfsgo.BlockChain
	TxPendingPool *xfsgo.TxPool
	number        int
}

type GetBlockByNumArgs struct {
	Number string `json:"number"`
}

type GetBlockByHashArgs struct {
	Hash string `json:"hash"`
}

type GetTxsByBlockNumArgs struct {
	Number string `json:"number"`
}
type GetTxbyBlockHashArgs struct {
	Hash string `json:"hash"`
}

type GetBalanceOfAddressArgs struct {
	Address string `json:"address"`
}

type GetTransactionArgs struct {
	Hash string `json:"hash"`
}

type GetReceiptByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBlockHeaderByNumberArgs struct {
	Number string `json:"number"`
	//Count  string `json:"count"`
}

type GetBlockHeaderByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBlocksByRangeArgs struct {
	From  string `json:"from"`
	Count string `json:"count"`
}

type GetBlocksArgs struct {
	Blocks string `json:"blocks"`
}

type ProgressBarArgs struct {
	Number int `json:"number"`
}

type GetBlockTxCountByHashArgs struct {
	Hash string `json:"hash"`
}

type GetBlockTxCountByNumArgs struct {
	Number string `json:"number"`
}

type GetBlockTxByHashAndIndexArgs struct {
	Hash  string `json:"hash"`
	Index int    `json:"index"`
}

type GetBlockTxByNumAndIndexArgs struct {
	Number string `json:"number"`
	Index  int    `json:"index"`
}
type GetBlockHashesArgs struct {
	Number string `json:"number"`
	Count  string `json:"count"`
}

func (handler *ChainAPIHandler) GetBlockByNumber(args GetBlockByNumArgs, resp **BlockResp) error {
	var last uint64
	if args.Number == "" {
		last = handler.BlockChain.CurrentBHeader().Height
	} else {
		number, ok := new(big.Int).SetString(args.Number, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
		last = number.Uint64()
	}
	gotBlock := handler.BlockChain.GetBlockByNumber(last)
	return coverBlock2Resp(gotBlock, resp)
}
func (handler *ChainAPIHandler) GetBlockHashes(args GetBlockHashesArgs, resp *[]common.Hash) error {
	start, _ := strconv.ParseUint(args.Number, 10, 64)
	count, _ := strconv.ParseUint(args.Count, 10, 64)
	last := handler.BlockChain.GetBlockByNumber(start + count - 1)
	if last == nil {
		last = handler.BlockChain.GetHead()
		count = last.Height() - start + 1
	}
	if last.Height() < start {
		return nil
	}
	hashes := []common.Hash{last.HeaderHash()}
	hashes = append(hashes, handler.BlockChain.GetBlockHashesFromHash(last.HeaderHash(), count-1)...)
	for i := 0; i < len(hashes)/2; i++ {
		hashes[i], hashes[len(hashes)-1-i] = hashes[len(hashes)-1-i], hashes[i]
	}
	*resp = hashes
	return nil
}

// func (handler *ChainAPIHandler) GetBlocksByNumber(args GetBlockByNumArgs, resp **BlocksResp) error {
// 	if args.Number == "" {
// 		return xfsgo.NewRPCError(-1006, "number not be empty")
// 	}
// 	number, ok := new(big.Int).SetString(args.Number, 0)
// 	if !ok {
// 		return xfsgo.NewRPCError(-1006, "number format error")
// 	}
// 	numbern := number.Uint64()
// 	gotBlocks := handler.BlockChain.GetBlocksFromNumber(numbern)
// 	return coverBlocks2Resp(gotBlocks, resp)
// }

func (handler *ChainAPIHandler) Head(_ EmptyArgs, resp **BlockHeaderResp) error {
	gotBlock := handler.BlockChain.GetHead()
	return coverBlockHeader2Resp(gotBlock, resp)
}

func (handler *ChainAPIHandler) GetBlockHeaderByNumber(args GetBlockHeaderByNumberArgs, resp **BlockHeaderResp) error {
	var last uint64
	if args.Number == "" {
		last = handler.BlockChain.CurrentBHeader().Height
	} else {
		number, ok := new(big.Int).SetString(args.Number, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
		last = number.Uint64()
	}
	gotBlock := handler.BlockChain.GetBlockByNumber(last)
	return coverBlockHeader2Resp(gotBlock, resp)
}

func (handler *ChainAPIHandler) GetBlockHeaderByHash(args GetBlockHeaderByHashArgs, resp **BlockHeaderResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	goBlock := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	return coverBlockHeader2Resp(goBlock, resp)
}

func (handler *ChainAPIHandler) GetBlockByHash(args GetBlockByHashArgs, resp **BlockResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	gotBlock := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	return coverBlock2Resp(gotBlock, resp)

}

func (handler *ChainAPIHandler) GetTxsByBlockNum(args GetTxsByBlockNumArgs, resp **TransactionsResp) error {
	var last uint64
	if args.Number == "" {
		last = handler.BlockChain.CurrentBHeader().Height
	} else {
		number, ok := new(big.Int).SetString(args.Number, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
		last = number.Uint64()
	}
	blk := handler.BlockChain.GetBlockByNumber(last)
	if blk == nil {
		return xfsgo.NewRPCError(-1006, "Not found block")
	}
	return coverTxs2Resp(blk.Transactions, resp)
}

func (handler *ChainAPIHandler) GetTxsByBlockHash(args GetTxbyBlockHashArgs, resp **TransactionsResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	blk := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	if blk == nil {
		return xfsgo.NewRPCError(-1006, "Not found block")
	}
	return coverTxs2Resp(blk.Transactions, resp)
}

func (handler *ChainAPIHandler) GetReceiptByHash(args GetReceiptByHashArgs, resp **ReceiptResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	dataReceipt := handler.BlockChain.GetReceiptByHash(common.Hex2Hash(args.Hash))
	if dataReceipt == nil {
		return xfsgo.NewRPCError(-1006, "Not found")
	}
	dataReceiptIndex := handler.BlockChain.GetReceiptByHashIndex(common.Hex2Hash(args.Hash))
	if dataReceiptIndex == nil {
		return xfsgo.NewRPCError(-1006, "Not found")
	}
	block := handler.BlockChain.GetBlockByHash(dataReceiptIndex.BlockHash)
	data := &ReceiptResp{
		Version:     dataReceipt.Version,
		Status:      dataReceipt.Status,
		TxHash:      dataReceipt.TxHash,
		GasUsed:     dataReceipt.GasUsed,
		BlockHeight: block.Height(),
		BlockHash:   dataReceiptIndex.BlockHash,
		BlockIndex:  dataReceiptIndex.BlockIndex,
		TxIndex:     dataReceiptIndex.Index,
	}

	return coverReceipt(data, resp)
}

func (handler *ChainAPIHandler) GetTransaction(args GetTransactionArgs, resp **TransactionResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	ID := common.Hex2Hash(args.Hash)
	data := handler.BlockChain.GetTransactionByTxHash(ID)
	return coverTx2Resp(data, resp)
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
func (handler *ChainAPIHandler) GetSyncStatus(_ EmptyArgs, resp *ChainStatusResp) error {
	current := handler.BlockChain.CurrentBHeader().Height
	origin, height := handler.BlockChain.Boundaries()

	var result *ChainStatusResp = nil
	if current < height {
		result = &ChainStatusResp{
			Status: true,
		}
	} else {
		result = &ChainStatusResp{
			Status: false,
		}
	}

	result.StartingBlock = new(big.Int).SetUint64(origin).Text(10)
	result.CurrentBlock = new(big.Int).SetUint64(current).Text(10)
	result.HighestBlock = new(big.Int).SetUint64(height).Text(10)

	*resp = *result

	return nil
}

func (handler *ChainAPIHandler) ExportBlocks(args GetBlocksByRangeArgs, resp *string) error {

	var numbersForm, numbersCount *big.Int
	var ok bool

	if args.From == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	} else {
		numbersForm, ok = new(big.Int).SetString(args.From, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
	}

	if args.Count == "" {
		blockHeight := handler.BlockChain.CurrentBHeader().Height
		numbersCount = new(big.Int).SetUint64(blockHeight)
	} else {
		numbersCount, ok = new(big.Int).SetString(args.Count, 0)
		if !ok {
			return xfsgo.NewRPCError(-1006, "string to big.Int error")
		}
	}

	if numbersCount.Uint64() == uint64(0) {
		b := handler.BlockChain.GetHead()
		numbersCount.SetUint64(b.Height())
	}

	if numbersForm.Uint64() >= numbersCount.Uint64() { // Export all
		b := handler.BlockChain.GetHead()
		numbersCount.SetUint64(b.Header.Height)
	}
	data := handler.BlockChain.GetBlocks(numbersForm.Uint64(), numbersCount.Uint64())
	encodeByte, err := json.Marshal(data)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	key := crypto.MD5Str(encodeByte)
	encryption, err := crypto.AesEncrypt(encodeByte, key)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	var bt bytes.Buffer
	bt.WriteString(key)
	bt.WriteString(encryption)
	respStr := bt.String()
	*resp = respStr
	return nil

}

func (handler *ChainAPIHandler) ImportBlock(args GetBlocksArgs, resp *string) error {
	if args.Blocks == "" {
		return xfsgo.NewRPCError(-1006, "to Blocks file path not be empty")
	}
	if len(args.Blocks[:]) < 32 {
		return xfsgo.NewRPCError(-1006, "No rules found for key")
	}
	key := args.Blocks[:32]

	decodeBuf := args.Blocks[32:]
	decryption, err := crypto.AesDecrypt(decodeBuf, key)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}

	blockChain := make([]*xfsgo.Block, 0)
	err = json.Unmarshal([]byte(decryption), &blockChain)
	if err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	handler.number = len(blockChain) - 1

	for _, item := range blockChain {
		if err = handler.BlockChain.InsertChain(item); err != nil {
			logrus.Errorf("Import block err: %v", err)
			continue
		}
	}
	*resp = "Import complete"
	return nil
}

func (handler *ChainAPIHandler) ProgressBar(_ EmptyArgs, resp *string) error {
	total := strconv.Itoa(handler.number)
	*resp = total
	return nil
}

func (handler *ChainAPIHandler) GetBlockTxCountByHash(args GetBlockTxCountByHashArgs, resp *int) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}
	gotBlock := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	result := len(gotBlock.Transactions)
	*resp = result
	return nil
}

func (handler *ChainAPIHandler) GetBlockTxCountByNum(args GetBlockTxCountByNumArgs, resp *int) error {
	if args.Number == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	number, ok := new(big.Int).SetString(args.Number, 0)
	if !ok {
		return xfsgo.NewRPCError(-1006, "number to big.Int Error")
	}
	gotBlock := handler.BlockChain.GetBlockByNumber(number.Uint64())
	result := len(gotBlock.Transactions)
	*resp = result
	return nil
}

func (handler *ChainAPIHandler) GetBlockTxByHashAndIndex(args GetBlockTxByHashAndIndexArgs, resp *TransactionResp) error {
	if args.Hash == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}

	if err := common.HashCalibrator(args.Hash); err != nil {
		return xfsgo.NewRPCErrorCause(-32001, err)
	}

	gotBlock := handler.BlockChain.GetBlockByHash(common.Hex2Hash(args.Hash))
	maxTxNumber := len(gotBlock.Transactions)
	if args.Index > maxTxNumber {
		return xfsgo.NewRPCError(-1006, "Block not found, the index transaction")
	}
	tx := gotBlock.Transactions[args.Index]
	return coverTx2Resp(tx, &resp)
}

func (handler *ChainAPIHandler) GetBlockTxByNumAndIndex(args GetBlockTxByNumAndIndexArgs, resp *TransactionResp) error {
	if args.Number == "" {
		return xfsgo.NewRPCError(-1006, "Parameter cannot be empty")
	}
	number, ok := new(big.Int).SetString(args.Number, 0)
	if !ok {
		return xfsgo.NewRPCError(-1006, "number to big.Int Error")
	}
	gotBlock := handler.BlockChain.GetBlockByNumber(number.Uint64())
	maxTxNumber := len(gotBlock.Transactions)
	if args.Index > maxTxNumber {
		return xfsgo.NewRPCError(-1006, "Block not found, the index transaction")
	}
	tx := gotBlock.Transactions[args.Index]
	return coverTx2Resp(tx, &resp)
}
