package configs

import (
	"flag"
)

var (
	RPCEndpoint          string
	WebsocketRPCEndpoint string
	FromBlock            int64
	CollectorsWorks      int
	MongoEndpoint        string
	MongoDatabase        string
	MongoCollection      string
)

func init() {
	rpcEndpoint := flag.String("rpcEndpoint", "", "JSON-RPC endpoint URL")
	wsEndpoint := flag.String("websocketRPCEndpoint", "", "WebSocket JSON-RPC endpoint URL")
	fromBlock := flag.Int64("fromBlock", 0, "Starting block number")
	collectorsWorks := flag.Int("collectorsWorks", 0, "Number of workers for collectors")
	mongoEndpoint := flag.String("mongoEndpoint", "", "MongoDB endpoint URL")
	mongoDatabase := flag.String("mongoDatabase", "", "MongoDB database name")
	mongoCollection := flag.String("mongoCollection", "", "MongoDB collection name")

	flag.Parse()

	RPCEndpoint = *rpcEndpoint
	WebsocketRPCEndpoint = *wsEndpoint
	FromBlock = *fromBlock
	CollectorsWorks = *collectorsWorks
	MongoEndpoint = *mongoEndpoint
	MongoDatabase = *mongoDatabase
	MongoCollection = *mongoCollection
}
