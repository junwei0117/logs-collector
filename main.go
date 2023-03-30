package main

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	routes "github.com/junwei0117/logs-collector/api/routers"
	"github.com/junwei0117/logs-collector/pkg/collectors"
	loggerCommon "github.com/junwei0117/logs-collector/pkg/common"
	"github.com/junwei0117/logs-collector/pkg/configs"
	"github.com/junwei0117/logs-collector/pkg/database"
	"github.com/junwei0117/logs-collector/pkg/logger"
	"github.com/junwei0117/logs-collector/pkg/subscriber"
)

func init() {
	logger.Init()
	err := database.ConnectToMongoDB()
	if err != nil {
		logger.Logger.Fatalf("[Database] Failed to connect to MongoDB: %v", err)
	}
}

func main() {
	logs, err := subscriber.SubscribeToTransferLogs()
	if err != nil {
		logger.Logger.Errorf("[Subscriber] Failed to subscribe to transfer events: %v", err)
	}

	go func() {
		for vLog := range logs {
			transferLog, err := loggerCommon.HandleTransferLogs(vLog)
			if err != nil {
				logger.Logger.Errorf("[Subscriber] Failed to handle transfer event: %v", err)
			}
			if transferLog != nil {
				logger.Logger.Infof("[Subscriber] Received transfer event: %v", transferLog)
			}
		}
	}()

	go func() {
		pastLogs, err := collectors.GetTransferLogs(configs.FromBlock)
		if err != nil {
			logger.Logger.Errorf("[Collector] Failed to get transfer events: %v", err)
		}

		logger.Logger.Infof("[Collector] Syncing %d past logs since block %v", len(pastLogs), configs.FromBlock)

		logChan := make(chan types.Log)
		var wg sync.WaitGroup
		wg.Add(configs.CollectorsWorks)

		for i := 0; i < configs.CollectorsWorks; i++ {
			go func() {
				defer wg.Done()
				for vLog := range logChan {
					transferLog, err := loggerCommon.HandleTransferLogs(vLog)
					if err != nil {
						logger.Logger.Errorf("[Collector] Failed to handle transfer event: %v", err)
					}
					if transferLog != nil {
						logger.Logger.Infof("[Collector] Received transfer event: %v", transferLog)
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
