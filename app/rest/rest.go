package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	// Extracted numeric query param, gets parsed into
	// unsigned integer
	parseNumber := func(number string) (uint64, error) {
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

			// Block hash based all tx retrieval request handler
			if strings.HasPrefix(hash, "0x") && len(hash) == 66 && tx == "yes" {
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

				_num, err := parseNumber(number)
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
			if strings.HasPrefix(hash, "0x") && len(hash) == 66 {
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

				_num, err := parseNumber(number)
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

		// Transaction fetch ( by query params ) request handler
		grp.GET("/transaction", func(c *gin.Context) {

			hash := c.Query("hash")

			// Simply returns single tx object, when queried using tx hash
			if strings.HasPrefix(hash, "0x") && len(hash) == 66 {
				if tx := db.GetTransactionByHash(_db, common.HexToHash(hash)); tx != nil {

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

			// block number range
			fromBlock := c.Query("fromBlock")
			toBlock := c.Query("toBlock")

			// block time span range
			fromTime := c.Query("fromTime")
			toTime := c.Query("toTime")

			// Contract deployer address
			deployer := c.Query("deployer")

			// account pair, in between this pair tx(s) to be extracted out in time span range/ block number range
			//
			// only single one can be supplied to enforce tx search
			// for incoming/ outgoing tx to & from an account
			fromAccount := c.Query("fromAccount")
			toAccount := c.Query("toAccount")

			// Account nonce, to be used for finding
			// tx, in combination with `fromAccount`
			nonce := c.Query("nonce")

			// Responds with tx sent from account with specified nonce
			if nonce != "" && strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42 {

				_nonce, err := parseNumber(nonce)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad account nonce",
					})
					return
				}

				if tx := db.GetTransactionFromAccountWithNonce(_db, common.HexToAddress(fromAccount), _nonce); tx != nil {
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

			// Responds with all contract creation tx(s) sent from specific account
			// ( i.e. deployer ) with in given block number range
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(deployer, "0x") && len(deployer) == 42 {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, 100)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetContractCreationTransactionsFromAccountByBlockNumberRange(_db, common.HexToAddress(deployer), _fromBlock, _toBlock); tx != nil {

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

			// Responds with all contract creation tx(s) sent from specific account
			// ( i.e. deployer ) with in given time frame
			if fromTime != "" && toTime != "" && strings.HasPrefix(deployer, "0x") && len(deployer) == 42 {

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, 600)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetContractCreationTransactionsFromAccountByBlockTimeRange(_db, common.HexToAddress(deployer), _fromTime, _toTime); tx != nil {

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

			// Given block number range & a pair of accounts, can find out all tx performed
			// between that pair, where `from` & `to` fields are fixed
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42 && strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42 {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, 100)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetTransactionsBetweenAccountsByBlockNumberRange(_db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _fromBlock, _toBlock); tx != nil {

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

			// Given block time range & a pair of accounts, can find out all tx performed
			// between that pair, where `from` & `to` fields are fixed
			if fromTime != "" && toTime != "" && strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42 && strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42 {

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, 600)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetTransactionsBetweenAccountsByBlockTimeRange(_db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _fromTime, _toTime); tx != nil {

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

			// Given block number range & account, can find out all tx performed
			// from account
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42 {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, 100)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetTransactionsFromAccountByBlockNumberRange(_db, common.HexToAddress(fromAccount), _fromBlock, _toBlock); tx != nil {

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

			// Given block mining time stamp range & account address, returns all outgoing tx
			// from this account in that given time span
			if fromTime != "" && toTime != "" && strings.HasPrefix(fromAccount, "0x") && len(fromAccount) == 42 {

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, 600)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetTransactionsFromAccountByBlockTimeRange(_db, common.HexToAddress(fromAccount), _fromTime, _toTime); tx != nil {

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

			// Given block number range & account address, returns all incoming tx
			// to this account in that block range
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42 {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, 100)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetTransactionsToAccountByBlockNumberRange(_db, common.HexToAddress(toAccount), _fromBlock, _toBlock); tx != nil {

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

			// Given block mining time stamp range & account address, returns all incoming tx
			// to this account in that given time span
			if fromTime != "" && toTime != "" && strings.HasPrefix(toAccount, "0x") && len(toAccount) == 42 {

				_fromTime, _toTime, err := rangeChecker(fromBlock, toBlock, 600)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetTransactionsToAccountByBlockTimeRange(_db, common.HexToAddress(toAccount), _fromTime, _toTime); tx != nil {

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

			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "Bad query param(s)",
			})

		})

		// Event(s) fetched by query params handler end point
		grp.GET("/event", func(c *gin.Context) {

			fromBlock := c.Query("fromBlock")
			toBlock := c.Query("toBlock")

			fromTime := c.Query("fromTime")
			toTime := c.Query("toTime")

			contract := c.Query("contract")

			blockHash := c.Query("blockHash")

			// Given blockhash, retrieves all events emitted by tx present in block
			if strings.HasPrefix(blockHash, "0x") && len(blockHash) == 66 {

				if event := db.GetEventsByBlockHash(_db, common.HexToHash(blockHash)); event != nil {

					if data := event.ToJSON(); data != nil {
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

			// Given block number range & contract address, finds out all events emitted by this contract
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(contract, "0x") && len(contract) == 42 {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, 10)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if event := db.GetEventsFromContractByBlockNumberRange(_db, common.HexToAddress(contract), _fromBlock, _toBlock); event != nil {

					if data := event.ToJSON(); data != nil {
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

			// Given block time span & contract address, returns a list of
			// events emitted by this contract during time span
			if fromTime != "" && toTime != "" && strings.HasPrefix(contract, "0x") && len(contract) == 42 {

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, 600)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if event := db.GetEventsFromContractByBlockTimeRange(_db, common.HexToAddress(contract), _fromTime, _toTime); event != nil {

					if data := event.ToJSON(); data != nil {
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
