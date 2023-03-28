package main

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
)

const (
	rpcEndpoint = "wss://ws.json-rpc.evm.testnet.shimmer.network"
)

var (
	contractAddresses = []common.Address{common.HexToAddress("0x00E77D6a7A56E8eD41B166EE7C3a887CC2FBc213")}
	erc20TransferSig  = []byte("Transfer(address,address,uint256)")
)

type LogTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	client, err := ethclient.DialContext(context.Background(), rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	logs, err := subscribeToTransferEvents(client)
	if err != nil {
		log.Fatalf("Failed to subscribe to transfer events: %v", err)
	}

	for vLog := range logs {
		err := handleTransferEvent(contractAbi, vLog)
		if err != nil {
			log.Printf("Failed to handle transfer event: %v", err)
		}
	}
}

func subscribeToTransferEvents(client *ethclient.Client) (<-chan types.Log, error) {
	topic := crypto.Keccak256Hash(erc20TransferSig)
	filter := ethereum.FilterQuery{
		Addresses: contractAddresses,
		Topics:    [][]common.Hash{{topic}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logs := make(chan types.Log)
	_, err := client.SubscribeFilterLogs(ctx, filter, logs)
	if err != nil {
		return nil, errors.New("failed to subscribe to transfer events: " + err.Error())
	}

	return logs, nil
}

func handleTransferEvent(contractAbi abi.ABI, vLog types.Log) error {
	var transferEvent LogTransfer

	err := contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
	if err != nil {
		return errors.New("failed to unpack transfer event data: " + err.Error())
	}

	transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
	transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

	log.Printf("Block Number: %d\n", vLog.BlockNumber)
	log.Printf("From: %s\n", transferEvent.From.Hex())
	log.Printf("To: %s\n", transferEvent.To.Hex())
	log.Printf("Tokens: %s\n", transferEvent.Value.String())
	log.Println("=====================================================")

	return nil
}
