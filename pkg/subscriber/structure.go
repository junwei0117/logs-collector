package subscriber

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
