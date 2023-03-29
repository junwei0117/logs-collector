package collectors

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/junwei0117/logs-collector/contracts/token"
	"github.com/junwei0117/logs-collector/pkg/configs"
	"github.com/junwei0117/logs-collector/pkg/database"
	"github.com/junwei0117/logs-collector/pkg/logger"
	"github.com/junwei0117/logs-collector/pkg/subscriber"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	Erc20TransferSig = []byte("Transfer(address,address,uint256)")
)

func GetTransferLogs(fromBlock int64, toBlock ...int64) ([]types.Log, error) {
	var endBlock int64
	if len(toBlock) > 0 {
		endBlock = toBlock[0]
	} else {
		client, err := ethclient.DialContext(context.Background(), configs.RPCEndpoint)
		if err != nil {
			return nil, err
		}
		block, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			return nil, err
		}
		endBlock = block.Number().Int64()
	}

	client, err := ethclient.DialContext(context.Background(), configs.WebsocketRPCEndpoint)
	if err != nil {
		return nil, err
	}

	topic := crypto.Keccak256Hash(Erc20TransferSig)

	var logs []types.Log

	for blockStart := fromBlock; blockStart <= endBlock; blockStart += 2000 {
		blockEnd := blockStart + 2000 - 1
		if blockEnd > endBlock {
			blockEnd = endBlock
		}

		filter := ethereum.FilterQuery{
			FromBlock: big.NewInt(blockStart),
			ToBlock:   big.NewInt(blockEnd),
			Topics:    [][]common.Hash{{topic}},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		blockLogs, err := client.FilterLogs(ctx, filter)
		if err != nil {
			return nil, errors.New("failed to subscribe to transfer events: " + err.Error())
		}

		logs = append(logs, blockLogs...)
	}

	return logs, nil
}

func HandleTransferLogs(vLog types.Log) error {
	if len(vLog.Data) == 0 {
		return nil
	}

	db, err := database.ConnectToMongoDB()
	if err != nil {
		return err
	}
	defer db.Client().Disconnect(context.Background())

	filter := bson.M{"txhash": vLog.TxHash}
	count, err := db.Collection(configs.MongoCollection).CountDocuments(context.Background(), filter)
	if err != nil {
		return errors.New("failed to count transfer event documents in MongoDB: " + err.Error())
	}
	if count > 0 {
		logger.Logger.Debugf("[Collector] Transfer event already exists in MongoDB: %+v", vLog.TxHash)
		return nil
	}

	transferEvent := &subscriber.TransferLog{}

	contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
	if err != nil {
		logger.Logger.Errorf("Failed to parse contract ABI: %v", err)
	}

	err = contractAbi.UnpackIntoInterface(transferEvent, "Transfer", vLog.Data)
	if err != nil {
		return errors.New("failed to unpack transfer event data: " + err.Error())
	}

	blockTimeStamp, err := GetBlockTimeStamp(vLog.BlockNumber)
	if err != nil {
		return err
	}

	transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
	transferEvent.ContractAddress = vLog.Address
	transferEvent.BlockNumber = vLog.BlockNumber
	transferEvent.BlockHash = vLog.BlockHash
	transferEvent.TxHash = vLog.TxHash
	transferEvent.TxIndex = vLog.TxIndex
	transferEvent.Index = vLog.Index
	transferEvent.BlockTimeStamp = blockTimeStamp

	logger.Logger.Infof("[Collector] Transfer event: %+v", transferEvent)

	_, err = db.Collection(configs.MongoCollection).InsertOne(context.Background(), transferEvent)
	if err != nil {
		return errors.New("failed to insert transfer event into MongoDB: " + err.Error())
	}

	return nil
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
		logger.Logger.Errorf("Failed to connect to Ethereum client: %v", err)
	}

	var block *types.Block
	var retries = 3
	var delay = time.Second * 1

	for retries > 0 {
		block, err = client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
		if err == nil {
			break
		}

		logger.Logger.Errorf("Failed to get block by number: %v. Retrying in %v...", err, delay)
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
