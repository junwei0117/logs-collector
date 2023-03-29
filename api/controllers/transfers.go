package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/junwei0117/logs-collector/pkg/database"
	"github.com/junwei0117/logs-collector/pkg/subscriber"
)

func GetTransfers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	db, err := database.ConnectToMongoDB()
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer db.Client().Disconnect(ctx)

	transfers := []*subscriber.TransferLog{}
	queryFilter := bson.M{}
	queryOptions := options.Find().SetSort(bson.M{"blockNumber": -1}).SetLimit(100)

	cursor, err := db.Collection(database.MongoCollection).Find(ctx, queryFilter, queryOptions)
	if err != nil {
		log.Printf("Failed to execute MongoDB query: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &transfers); err != nil {
		log.Printf("Failed to parse MongoDB result: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, transfers)
}

func GetTransfersCount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	db, err := database.ConnectToMongoDB()
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer db.Client().Disconnect(ctx)

	fromBlockStr := c.Query("fromBlock")
	toBlockStr := c.Query("toBlock")

	fromBlock, err := strconv.ParseUint(fromBlockStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fromBlock parameter"})
		return
	}

	toBlock, err := strconv.ParseUint(toBlockStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid toBlock parameter"})
		return
	}

	filter := bson.M{
		"blocknumber": bson.M{
			"$gte": fromBlock,
			"$lte": toBlock,
		},
	}

	count, err := db.Collection(database.MongoCollection).CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("Failed to execute MongoDB query: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
