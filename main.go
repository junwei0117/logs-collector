package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	routes "github.com/junwei0117/logs-collector/api/routers"
	"github.com/junwei0117/logs-collector/pkg/collectors"
	"github.com/junwei0117/logs-collector/pkg/subscriber"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logs, err := subscriber.SubscribeToTransferLogs()
	if err != nil {
		log.Printf("Failed to subscribe to transfer events: %v", err)
	}

	go func() {
		for vLog := range logs {
			err := subscriber.HandleTransferLogs(vLog)
			if err != nil {
				log.Printf("Failed to handle transfer event: %v", err)
			}
		}
	}()

	go func() {
		pastLogs, err := collectors.GetTransferLogs(24000)
		if err != nil {
			log.Printf("Failed to get transfer events: %v", err)
		}

		numWorkers := 2
		logChan := make(chan types.Log)
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for vLog := range logChan {
					err := collectors.HandleTransferLogs(vLog)
					if err != nil {
						log.Printf("Failed to handle transfer event: %v", err)
					}
				}
			}()
		}

		for _, vLog := range pastLogs {
			logChan <- vLog
		}

		close(logChan)

		wg.Wait()

		log.Println("[Collector] Done syncing past logs")
	}()

	r := routes.SetUpRouters()
	r.Run(fmt.Sprintf(":%s", "8080"))
}
