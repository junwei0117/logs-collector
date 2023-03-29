package main

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	routes "github.com/junwei0117/logs-collector/api/routers"
	"github.com/junwei0117/logs-collector/pkg/collectors"
	"github.com/junwei0117/logs-collector/pkg/logger"
	"github.com/junwei0117/logs-collector/pkg/subscriber"
)

var (
	fromBlock = int64(24000)
)

func init() {
	logger.Init()
}

func main() {

	logs, err := subscriber.SubscribeToTransferLogs()
	if err != nil {
		logger.Logger.Errorf("Failed to subscribe to transfer events: %v", err)
	}

	go func() {
		for vLog := range logs {
			err := subscriber.HandleTransferLogs(vLog)
			if err != nil {
				logger.Logger.Errorf("Failed to handle transfer event: %v", err)
			}
		}
	}()

	go func() {
		pastLogs, err := collectors.GetTransferLogs(fromBlock)
		if err != nil {
			logger.Logger.Errorf("Failed to get transfer events: %v", err)
		}

		logger.Logger.Infof("[Collector] Syncing %d past logs since block %v", len(pastLogs), fromBlock)

		numWorkers := 5
		logChan := make(chan types.Log)
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for vLog := range logChan {
					err := collectors.HandleTransferLogs(vLog)
					if err != nil {
						logger.Logger.Errorf("Failed to handle transfer event: %v", err)
					}
				}
			}()
		}

		for _, vLog := range pastLogs {
			logChan <- vLog
		}

		close(logChan)

		wg.Wait()

		logger.Logger.Infof("[Collector] Done syncing past logs")
	}()

	r := routes.SetUpRouters()
	r.Run(fmt.Sprintf(":%s", "8080"))
}
