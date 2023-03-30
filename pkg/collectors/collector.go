package collectors

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/junwei0117/logs-collector/pkg/configs"
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
			return nil, err
		}

		logs = append(logs, blockLogs...)
	}

	return logs, nil
}
