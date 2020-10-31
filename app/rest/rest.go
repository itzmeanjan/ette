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

	// Extracted from, to field of range based block query ( using block numbers/ time stamps )
	// gets parsed into unsigned integers
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
			return 0, 0, errors.New("Range too long")
		}

		return _from, _to, nil
	}

	// Extracted block number from URL query string, gets parsed into
	// unsigned integer
	parseBlockNumber := func(number string) (uint64, error) {
		_num, err := strconv.ParseUint(number, 10, 64)
		if err != nil {
			return 0, errors.New("Failed to parse integer")
		}

		return _num, nil
	}

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

			// Given block number, finds out all tx(s) present in that block
			if number != "" && tx == "yes" {

				_num, err := parseBlockNumber(number)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number",
					})
					return
				}

				if tx := db.GetTransactionsByBlockNumber(_db, _num); tx != nil {
					if data := tx.ToJSON(); data != nil {
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

				_num, err := parseBlockNumber(number)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number",
					})
					return
				}

				if block := db.GetBlockByNumber(_db, _num); block != nil {

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

			// Block number range based query
			// At max 10 blocks at a time to be returned
			fromBlock := c.Query("fromBlock")
			toBlock := c.Query("toBlock")

			if fromBlock != "" && toBlock != "" {

				_from, _to, err := rangeChecker(fromBlock, toBlock, 10)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if blocks := db.GetBlocksByNumberRange(_db, _from, _to); blocks != nil {

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

				_from, _to, err := rangeChecker(fromTime, toTime, 60)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if blocks := db.GetBlocksByTimeRange(_db, _from, _to); blocks != nil {

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
