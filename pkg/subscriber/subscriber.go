package subscriber

import (
	"context"
	"errors"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/junwei0117/logs-collector/contracts/token"
	"github.com/junwei0117/logs-collector/pkg/database"
)

func SubscribeToTransferEvents() (<-chan types.Log, error) {
	client, err := ethclient.DialContext(context.Background(), RPCEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	topic := crypto.Keccak256Hash(Erc20TransferSig)

	filter := ethereum.FilterQuery{
		Topics: [][]common.Hash{{topic}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logs := make(chan types.Log)
	_, err = client.SubscribeFilterLogs(ctx, filter, logs)
	if err != nil {
		return nil, errors.New("failed to subscribe to transfer events: " + err.Error())
	}

	return logs, nil
}

func HandleTransferEvent(vLog types.Log) error {
	if len(vLog.Data) == 0 {
		return nil
	}

	db, err := database.ConnectToMongoDB()
	if err != nil {
		return err
	}
	defer db.Client().Disconnect(context.Background())

	transferEvent := &TransferLog{}

	contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
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

	_, err = db.Collection(database.MongoCollection).InsertOne(context.Background(), transferEvent)
	if err != nil {
		return errors.New("failed to insert transfer event into MongoDB: " + err.Error())
	}

	return nil
}

func GetBlockTimeStamp(blockNumber uint64) (uint64, error) {
	client, err := ethclient.DialContext(context.Background(), RPCEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	block, err := client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
	if err != nil {
		return 0, errors.New("failed to get block by number: " + err.Error())
	}

	return block.Time(), nil
}
