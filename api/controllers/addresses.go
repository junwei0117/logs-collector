package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	loggerCommon "github.com/junwei0117/logs-collector/pkg/common"
	"github.com/junwei0117/logs-collector/pkg/configs"
	"github.com/junwei0117/logs-collector/pkg/database"
	"github.com/junwei0117/logs-collector/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAddresses(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	db, err := database.GetDB()
	if err != nil {
		logger.Logger.Errorf("Failed to connect to MongoDB: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer db.Client().Disconnect(ctx)

	transfers := []*loggerCommon.TransferLog{}

	fromBlockStr := c.Query("from_block")
	toBlockStr := c.Query("to_block")
	addressStr := c.Param("address")
	address := common.HexToAddress(addressStr)
	fromTimeStr := c.Query("from_time")
	toTimeStr := c.Query("to_time")

	var queryFilter bson.M

	if fromBlockStr != "" || toBlockStr != "" {
		if fromTimeStr != "" || toTimeStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot use both block and time filters"})
			return
		}

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

		if fromBlockStr != "" || toBlockStr != "" {
			blockFilter := bson.M{}
			if fromBlockStr != "" && toBlockStr != "" {
				blockFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock, "$lte": toBlock}}
			} else if fromBlockStr != "" {
				blockFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock}}
			} else if toBlockStr != "" {
				blockFilter = bson.M{"blocknumber": bson.M{"$lte": toBlock}}
			}

			filter := bson.M{
				"$and": []bson.M{
					{
						"$or": []bson.M{
							{"from": address},
							{"to": address},
							{"contractaddress": address},
						},
					},
					blockFilter,
				},
			}

			queryFilter = filter
		}
	} else if fromTimeStr != "" || toTimeStr != "" {
		fromTime, err := strconv.ParseUint(fromTimeStr, 10, 64)
		if err != nil && fromTimeStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fromTime parameter"})
			return
		}

		toTime, err := strconv.ParseUint(toTimeStr, 10, 64)
		if err != nil && toTimeStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid toTime parameter"})
			return
		}

		if fromTimeStr != "" || toTimeStr != "" {
			blockFilter := bson.M{}
			if fromTimeStr != "" && toTimeStr != "" {
				blockFilter = bson.M{"blocktimestamp": bson.M{"$gte": fromTime, "$lte": toTime}}
			} else if fromTimeStr != "" {
				blockFilter = bson.M{"blocktimestamp": bson.M{"$gte": fromTime}}
			} else if toTimeStr != "" {
				blockFilter = bson.M{"blocktimestamp": bson.M{"$lte": toTime}}
			}

			filter := bson.M{
				"$and": []bson.M{
					{
						"$or": []bson.M{
							{"from": address},
							{"to": address},
							{"contractaddress": address},
						},
					},
					blockFilter,
				},
			}

			queryFilter = filter
		}
	} else {
		queryFilter = bson.M{
			"$or": []bson.M{
				{"from": address},
				{"to": address},
				{"contractaddress": address},
			},
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	offset := (page - 1) * pageSize
	limit := pageSize

	queryOptions := options.Find().SetSort(bson.M{"blocknumber": 1}).SetSkip(int64(offset)).SetLimit(int64(limit))

	cursor, err := db.Collection(configs.MongoCollection).Find(ctx, queryFilter, queryOptions)
	if err != nil {
		logger.Logger.Errorf("Failed to execute MongoDB query: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &transfers); err != nil {
		logger.Logger.Errorf("Failed to parse MongoDB result: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, transfers)
}

func GetAddressesCount(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	db, err := database.GetDB()
	if err != nil {
		logger.Logger.Errorf("Failed to connect to MongoDB: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer db.Client().Disconnect(ctx)

	fromBlockStr := c.Query("from_block")
	toBlockStr := c.Query("to_block")
	addressStr := c.Param("address")
	address := common.HexToAddress(addressStr)
	fromTimeStr := c.Query("from_time")
	toTimeStr := c.Query("to_time")

	var queryFilter bson.M

	if fromBlockStr != "" || toBlockStr != "" {
		if fromTimeStr != "" || toTimeStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot use both block and time filters"})
			return
		}

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

		if fromBlockStr != "" || toBlockStr != "" {
			blockFilter := bson.M{}
			if fromBlockStr != "" && toBlockStr != "" {
				blockFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock, "$lte": toBlock}}
			} else if fromBlockStr != "" {
				blockFilter = bson.M{"blocknumber": bson.M{"$gte": fromBlock}}
			} else if toBlockStr != "" {
				blockFilter = bson.M{"blocknumber": bson.M{"$lte": toBlock}}
			}

			filter := bson.M{
				"$and": []bson.M{
					{
						"$or": []bson.M{
							{"from": address},
							{"to": address},
							{"contractaddress": address},
						},
					},
					blockFilter,
				},
			}

			queryFilter = filter
		}
	} else if fromTimeStr != "" || toTimeStr != "" {
		fromTime, err := strconv.ParseUint(fromTimeStr, 10, 64)
		if err != nil && fromTimeStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid fromTime parameter"})
			return
		}

		toTime, err := strconv.ParseUint(toTimeStr, 10, 64)
		if err != nil && toTimeStr != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid toTime parameter"})
			return
		}

		if fromTimeStr != "" || toTimeStr != "" {
			blockFilter := bson.M{}
			if fromTimeStr != "" && toTimeStr != "" {
				blockFilter = bson.M{"blocktimestamp": bson.M{"$gte": fromTime, "$lte": toTime}}
			} else if fromTimeStr != "" {
				blockFilter = bson.M{"blocktimestamp": bson.M{"$gte": fromTime}}
			} else if toTimeStr != "" {
				blockFilter = bson.M{"blocktimestamp": bson.M{"$lte": toTime}}
			}

			filter := bson.M{
				"$and": []bson.M{
					{
						"$or": []bson.M{
							{"from": address},
							{"to": address},
							{"contractaddress": address},
						},
					},
					blockFilter,
				},
			}

			queryFilter = filter
		}
	} else {
		queryFilter = bson.M{
			"$or": []bson.M{
				{"from": address},
				{"to": address},
				{"contractaddress": address},
			},
		}
	}

	count, err := db.Collection(configs.MongoCollection).CountDocuments(ctx, queryFilter)
	if err != nil {
		logger.Logger.Errorf("Failed to execute MongoDB query: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
