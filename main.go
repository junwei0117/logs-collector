package main

import (
	"fmt"
	"log"

	routes "github.com/junwei0117/logs-collector/api/routers"
	"github.com/junwei0117/logs-collector/pkg/subscriber"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	logs, err := subscriber.SubscribeToTransferEvents()
	if err != nil {
		log.Fatalf("Failed to subscribe to transfer events: %v", err)
	}

	go func() {
		for vLog := range logs {
			err := subscriber.HandleTransferEvent(vLog)
			if err != nil {
				log.Printf("Failed to handle transfer event: %v", err)
			}
		}
	}()

	r := routes.SetUpRouters()
	r.Run(fmt.Sprintf(":%s", "8080"))
}
