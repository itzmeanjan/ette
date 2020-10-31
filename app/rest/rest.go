package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	"gorm.io/gorm"
)

// RunHTTPServer - Holds definition for all REST API(s) to be exposed
func RunHTTPServer(_db *gorm.DB, _lock *sync.Mutex, _synced *d.SyncState) {
	router := gin.Default()

	grp := router.Group("/v1")

	{
		// For checking whether `ette` has synced upto blockchain latest state or not
		grp.GET("/synced", func(c *gin.Context) {

			_lock.Lock()
			defer _lock.Unlock()

			c.JSON(http.StatusOK, gin.H{
				"synced": _synced.Synced,
			})

		})

		// Query block data using block hash/ number/ block number range ( 10 at max )
		grp.GET("/block", func(c *gin.Context) {

			hash := c.Query("hash")
			number := c.Query("number")
			tx := c.Query("tx")

			// Block hash based all tx in block retrieval request handler
			if hash != "" && tx == "yes" {
				if tx := db.GetTransactionsByBlockHash(_db, common.HexToHash(hash)); tx != nil {
					if data := tx.ToJSON(); data != nil {
						c.Data(http.StatusOK, "application/json", tx.ToJSON())
						return
					}

					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "JSON encoding failed",
					})
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return
			}

			// Given block number, finds out all tx(s) present in that block
			if number != "" && tx == "yes" {
				if tx := db.GetTransactionsByBlockNumber(_db, number); tx != nil {
					if data := tx.ToJSON(); data != nil {
						c.Data(http.StatusOK, "application/json", tx.ToJSON())
						return
					}

					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "JSON encoding failed",
					})
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return
			}

			// Block hash based single block retrieval request handler
			if hash != "" {
				if block := db.GetBlockByHash(_db, common.HexToHash(hash)); block != nil {

					if data := block.ToJSON(); data != nil {
						c.Data(http.StatusOK, "application/json", data)
						return
					}

					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "JSON encoding failed",
					})
					return

				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return
			}

			// Block number based single block retrieval request handler
			if number != "" {
				if block := db.GetBlockByNumber(_db, number); block != nil {

					if data := block.ToJSON(); data != nil {
						c.Data(http.StatusOK, "application/json", data)
						return
					}

					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "JSON encoding failed",
					})
					return

				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return
			}

			rangeChecker := func(from string, to string, limit uint64) (uint64, uint64, error) {
				_from, err := strconv.ParseUint(from, 10, 64)
				if err != nil {
					return 0, 0, errors.New("Failed to parse integer")
				}

				_to, err := strconv.ParseUint(to, 10, 64)
				if err != nil {
					return 0, 0, errors.New("Failed to parse integer")
				}

				if !(_to-_from < limit) {
					return _from, _to, nil
				}

				return 0, 0, errors.New("Failed to parse integer")
			}

			// Block number range based query
			// At max 10 blocks at a time to be returned
			fromBlock := c.Query("fromBlock")
			toBlock := c.Query("toBlock")

			if fromBlock != "" && toBlock != "" {
				if blocks := db.GetBlocksByNumberRange(_db, fromBlock, toBlock); blocks != nil {

					if data := blocks.ToJSON(); data != nil {
						c.Data(http.StatusOK, "application/json", data)
						return
					}

					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "JSON encoding failed",
					})
					return

				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return
			}

			// Query blocks by timestamp range, at max 60 seconds of timestamp
			// can be mentioned, otherwise request to be rejected
			fromTime := c.Query("fromTime")
			toTime := c.Query("toTime")

			if fromTime != "" && toTime != "" {
				if blocks := db.GetBlocksByTimeRange(_db, fromTime, toTime); blocks != nil {

					if data := blocks.ToJSON(); data != nil {
						c.Data(http.StatusOK, "application/json", data)
						return
					}

					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": "JSON encoding failed",
					})
					return

				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "Bad query param(s)",
			})

		})
	}

	router.Run(fmt.Sprintf(":%s", cfg.Get("PORT")))
}
