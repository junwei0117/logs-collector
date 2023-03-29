package collectors

import (
	"context"
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetTransferLogs(fromBlock int64, toBlock int64) ([]types.Log, error) {
	client, err := ethclient.DialContext(context.Background(), RPCEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	topic := crypto.Keccak256Hash(Erc20TransferSig)

	var logs []types.Log

	for blockStart := fromBlock; blockStart < toBlock; blockStart += 2000 {
		blockEnd := blockStart + 2000 - 1
		if blockEnd > toBlock {
			blockEnd = toBlock
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
