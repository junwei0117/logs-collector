package common

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/junwei0117/logs-collector/contracts/token"
	"github.com/junwei0117/logs-collector/pkg/configs"
	"github.com/junwei0117/logs-collector/pkg/database"
	"github.com/junwei0117/logs-collector/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type TransferLog struct {
	From            common.Address `json:"from"`
	To              common.Address `json:"to"`
	Value           *big.Int       `json:"value"`
	ContractAddress common.Address `json:"contractAddress"`
	BlockNumber     uint64         `json:"blockNumber"`
	BlockHash       common.Hash    `json:"blockHash"`
	TxHash          common.Hash    `json:"txHash"`
	TxIndex         uint           `json:"txIndex"`
	Index           uint           `json:"index"`
	BlockTimeStamp  uint64         `json:"blockTimeStamp"`
}

var blockTimeCache = struct {
	sync.Mutex
	m map[uint64]uint64
}{
	m: make(map[uint64]uint64),
}

func GetBlockTimeStamp(blockNumber uint64) (uint64, error) {
	blockTimeCache.Lock()
	defer blockTimeCache.Unlock()

	if timestamp, ok := blockTimeCache.m[blockNumber]; ok {
		return timestamp, nil
	}

	client, err := ethclient.DialContext(context.Background(), configs.RPCEndpoint)
	if err != nil {
		return 0, err
	}

	var block *types.Block
	var retries = 3
	var delay = time.Second * 1

	for retries > 0 {
		block, err = client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
		if err == nil {
			break
		}

		logger.Logger.Warnf("failed to get block by number: %v. retrying in %v...", err, delay)
		time.Sleep(delay)
		retries--
	}

	if err != nil {
		return 0, errors.New("failed to get block by number: " + err.Error())
	}

	blockTime := block.Time()
	blockTimeCache.m[blockNumber] = blockTime

	return blockTime, nil
}

func HandleTransferLogs(vLog types.Log) (*TransferLog, error) {
	if len(vLog.Data) == 0 {
		return nil, nil
	}

	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	filter := bson.M{"txhash": vLog.TxHash, "index": vLog.Index}
	count, err := db.Collection(configs.MongoCollection).CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		logger.Logger.Debugf("transfer event already exists in MongoDB: %+v", vLog.TxHash)
		return nil, nil
	}

	transferLog := &TransferLog{}

	contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
	if err != nil {
		return nil, err
	}

	err = contractAbi.UnpackIntoInterface(transferLog, "Transfer", vLog.Data)
	if err != nil {
		return nil, err
	}

	blockTimeStamp, err := GetBlockTimeStamp(vLog.BlockNumber)
	if err != nil {
		return nil, err
	}

	transferLog.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferLog.To = common.HexToAddress(vLog.Topics[2].Hex())
	transferLog.ContractAddress = vLog.Address
	transferLog.BlockNumber = vLog.BlockNumber
	transferLog.BlockHash = vLog.BlockHash
	transferLog.TxHash = vLog.TxHash
	transferLog.TxIndex = vLog.TxIndex
	transferLog.Index = vLog.Index
	transferLog.BlockTimeStamp = blockTimeStamp

	_, err = db.Collection(configs.MongoCollection).InsertOne(context.Background(), transferLog)
	if err != nil {
		return nil, err
	}

	return transferLog, nil
}
