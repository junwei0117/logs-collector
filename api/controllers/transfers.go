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

	fromBlockStr := c.Query("from_block")
	toBlockStr := c.Query("to_block")

	var queryFilter bson.M

	if fromBlockStr != "" || toBlockStr != "" {
		fromBlock, err := strconv.ParseUint(fromBlockStr, 10, 64)
		if err != nil && fromBlockStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fromBlock parameter"})
			return
		}

		toBlock, err := strconv.ParseUint(toBlockStr, 10, 64)
		if err != nil && toBlockStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid toBlock parameter"})
			return
		}

		if fromBlockStr != "" && toBlockStr != "" {
			queryFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock, "$lte": toBlock}}
		} else if fromBlockStr != "" {
			queryFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock}}
		} else if toBlockStr != "" {
			queryFilter = bson.M{"blocknumber": bson.M{"$lte": toBlock}}
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	offset := (page - 1) * pageSize
	limit := pageSize

	queryOptions := options.Find().SetSort(bson.M{"blocknumber": 1}).SetSkip(int64(offset)).SetLimit(int64(limit))

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

	fromBlockStr := c.Query("from_block")
	toBlockStr := c.Query("to_block")

	var queryFilter bson.M

	if fromBlockStr != "" || toBlockStr != "" {
		fromBlock, err := strconv.ParseUint(fromBlockStr, 10, 64)
		if err != nil && fromBlockStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fromBlock parameter"})
			return
		}

		toBlock, err := strconv.ParseUint(toBlockStr, 10, 64)
		if err != nil && toBlockStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid toBlock parameter"})
			return
		}

		if fromBlockStr != "" && toBlockStr != "" {
			queryFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock, "$lte": toBlock}}
		} else if fromBlockStr != "" {
			queryFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock}}
		} else if toBlockStr != "" {
			queryFilter = bson.M{"blocknumber": bson.M{"$lte": toBlock}}
		}
	}

	count, err := db.Collection(database.MongoCollection).CountDocuments(ctx, queryFilter)
	if err != nil {
		log.Printf("Failed to execute MongoDB query: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
