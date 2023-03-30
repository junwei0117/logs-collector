package subscriber

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/junwei0117/logs-collector/pkg/configs"
	"github.com/junwei0117/logs-collector/pkg/logger"
)

var (
	Erc20TransferSig = []byte("Transfer(address,address,uint256)")
)

func SubscribeToTransferLogs() (<-chan types.Log, error) {
	client, err := ethclient.DialContext(context.Background(), configs.WebsocketRPCEndpoint)
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Ethereum client: %v", err)
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
