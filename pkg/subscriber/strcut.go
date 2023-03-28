package subscriber

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TransferLog struct {
	From            common.Address
	To              common.Address
	Value           *big.Int
	ContractAddress common.Address
	BlockNumber     uint64
	BlockHash       common.Hash
	TxHash          common.Hash
	TxIndex         uint
	Index           uint
}
