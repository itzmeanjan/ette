package rest

import (
	"fmt"
	"net/http"
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

			// Block hash based single block retrieval query
			hash := c.Query("hash")

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

				c.JSON(http.StatusNoContent, gin.H{
					"msg": "Bad block hash",
				})
				return
			}

			// Block number based single block retrieval query
			number := c.Query("number")

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

				c.JSON(http.StatusNoContent, gin.H{
					"msg": "Bad block hash",
				})
				return
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

				c.JSON(http.StatusNoContent, gin.H{
					"msg": "Bad block number range",
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

				c.JSON(http.StatusNoContent, gin.H{
					"msg": "Bad block timestamp range",
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
