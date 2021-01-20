package rest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-contrib/cors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
	ps "github.com/itzmeanjan/ette/app/pubsub"
	"github.com/itzmeanjan/ette/app/rest/graph/generated"
	"gorm.io/gorm"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/itzmeanjan/ette/app/rest/graph"
)

// RunHTTPServer - Holds definition for all REST API(s) to be exposed
func RunHTTPServer(_db *gorm.DB, _status *d.StatusHolder, _redisClient *redis.Client) {

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

	// Return topics to be used for finding out events in hex form
	// topics are extracted out from query params in string form
	getTopics := func(topics ...string) []common.Hash {
		if topics[0] != "" && topics[1] != "" && topics[2] != "" && topics[3] != "" {
			return []common.Hash{common.HexToHash(topics[0]), common.HexToHash(topics[1]), common.HexToHash(topics[2]), common.HexToHash(topics[3])}
		}

		if topics[0] != "" && topics[1] != "" && topics[2] != "" {
			return []common.Hash{common.HexToHash(topics[0]), common.HexToHash(topics[1]), common.HexToHash(topics[2])}
		}

		if topics[0] != "" && topics[1] != "" {
			return []common.Hash{common.HexToHash(topics[0]), common.HexToHash(topics[1])}
		}

		if topics[0] != "" {
			return []common.Hash{common.HexToHash(topics[0])}
		}

		return nil
	}

	respondWithJSON := func(data []byte, c *gin.Context) {

		uri := c.Request.RequestURI

		// API key based client identification
		//
		// Data delivery being logged for implementing rate limiting
		user := db.GetUserFromAPIKey(_db, c.GetHeader("APIKey"))
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "Bad API Key",
			})
			return
		}

		if data != nil {
			c.Data(http.StatusOK, "application/json", data)

			switch {
			case strings.HasPrefix(uri, "/v1/block"):
				db.PutDataDeliveryInfo(_db, user.Address, "/v1/block", uint64(len(data)))
			case strings.HasPrefix(uri, "/v1/transaction"):
				db.PutDataDeliveryInfo(_db, user.Address, "/v1/transaction", uint64(len(data)))
			case strings.HasPrefix(uri, "/v1/event"):
				db.PutDataDeliveryInfo(_db, user.Address, "/v1/event", uint64(len(data)))
			}

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "JSON encoding failed",
		})
		return
	}

	// Validates sessionId, which is passed as cookie for
	// `/v1/dashboard/*` endpoints
	//
	// This cookie is set when login is performed
	validateSessionID := func(c *gin.Context) string {
		sessionID, err := c.Cookie("SessionID")
		if err == http.ErrNoCookie {
			return ""
		}

		address, err := _redisClient.Get(context.Background(), sessionID).Result()
		if err != nil {
			return ""
		}

		return address
	}

	// For any historical query request
	// APIKey needs to be delivered in header
	//
	// headers: { 'APIKey': '0x...' }
	// Which is checked against database & responded
	// accordingly
	validateAPIKey := func(c *gin.Context) {

		apiKey := c.GetHeader("APIKey")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "API Key Required",
			})
			return
		}

		user := db.GetUserFromAPIKey(_db, apiKey)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "Bad API Key",
			})
			return
		}

		// Checking if user has kept this APIKey enabled or not
		if !user.Enabled {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "Bad API Key",
			})
			return
		}

		// Checking if user has crossed allowed rate limit or not
		// If yes, we're dropping request
		if !db.IsUnderRateLimit(_db, user.Address) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"msg": "Crossed Allowed Rate Limit",
			})
			return
		}

		c.Next()

	}

	// Checking whether this `ette` instance support
	// historical data query or not
	checkEtteHistoricalMode := func(c *gin.Context) {
		if !(cfg.Get("EtteMode") == "1" || cfg.Get("EtteMode") == "3") {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"msg": "Disabled Feature",
			})
			return
		}

		c.Next()
	}

	// Checking whether this `ette` instance support
	// real-time data delivery or not, if not letting client know
	// about it & closing connection
	checkEtteRealTimeMode := func(conn *websocket.Conn) bool {
		if !(cfg.Get("EtteMode") == "2" || cfg.Get("EtteMode") == "3") {
			if err := conn.WriteJSON(&ps.SubscriptionResponse{Code: 0, Message: "Disabled Feature"}); err != nil {
				log.Printf("[!] Failed to write message : %s\n", err.Error())
			}
			return false
		}

		return true
	}

	// Checking if user has asked to run webserver in production mode or not
	checkIfInProduction := func() bool {
		return strings.ToLower(cfg.Get("Production")) == "yes"
	}

	// Running in production/ debug mode depending upon
	// user choice specified in .env file
	if checkIfInProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()

	// enabled cors
	router.Use(cors.Default())

	router.HTMLRender = ginview.New(goview.Config{
		Root:         "./views",
		Master:       "layouts/master",
		Extension:    ".html",
		DisableCache: true,
	})

	// Delivering `ette` icon to be shown in web UI
	router.GET("/favicon.ico", func(c *gin.Context) {
		data, err := ioutil.ReadFile("./favicon.ico")
		if err != nil {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		c.Data(http.StatusOK, "image/x-icon", data)
	})

	grp := router.Group("/v1")

	{

		grp.POST("/login", func(c *gin.Context) {

			var payload d.AuthPayload

			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "Bad Authentication Payload",
				})
				return
			}

			signer := payload.RecoverSigner()

			if !payload.VerifySignature(signer) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg": "Verification Failed",
				})
				return
			}

			if payload.HasExpired(30) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg": "Signature Expired",
				})
				return
			}

			if _, err := _redisClient.Set(context.Background(), payload.Signature, payload.Message.Address.Hex(), time.Duration(3600)*time.Second).Result(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Something went wrong",
				})
				return
			}

			c.SetCookie("SessionID", payload.Signature, 3600, "/v1/dashboard", cfg.Get("Domain"), false, false)

			if payload.IsAdmin(signer) {
				c.JSON(http.StatusOK, gin.H{
					"msg": "Success",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"msg": "Success",
			})

		})

		grp.GET("/login", func(c *gin.Context) {

			c.HTML(http.StatusOK, "index", gin.H{
				"title": "ette: Ethereum Blockchain Indexing Engine",
			})

		})

		grp.GET("/dashboard", func(c *gin.Context) {

			address := validateSessionID(c)
			if address == "" {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			c.HTML(http.StatusOK, "dashboard", gin.H{
				"title": "ette: Ethereum Blockchain Indexing Engine",
			})

		})

		grp.GET("/dashboard/apps", func(c *gin.Context) {

			address := validateSessionID(c)
			if address == "" {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			if apps := db.GetAppsByUserAddress(_db, common.HexToAddress(address)); apps != nil {
				c.JSON(http.StatusOK, gin.H{
					"apps": apps,
				})
				return
			}

			c.JSON(http.StatusNoContent, gin.H{
				"msg": "No apps created yet",
			})

		})

		grp.POST("/dashboard/newApp", func(c *gin.Context) {

			address := validateSessionID(c)
			if address == "" {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			var payload d.AuthPayload

			if err := c.ShouldBindJSON(&payload); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "Bad Authentication Payload",
				})
				return
			}

			signer := payload.RecoverSigner()

			if !payload.VerifySignature(signer) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg": "Verification Failed",
				})
				return
			}

			if payload.HasExpired(30) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg": "Signature Expired",
				})
				return
			}

			if common.HexToAddress(address) != common.BytesToAddress(signer) {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			if !db.RegisterNewApp(_db, common.HexToAddress(address)) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Failed to register app",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"msg": "Success",
			})

		})

		grp.POST("/dashboard/toggleApp", func(c *gin.Context) {

			address := validateSessionID(c)
			if address == "" {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			var apiKey d.APIKey

			if err := c.ShouldBindJSON(&apiKey); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "Bad APIKey Payload",
				})
				return
			}

			if !db.ToggleAPIKeyState(_db, apiKey.APIKey.Hex()) {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Failed to toggle app state",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"msg": "Success",
			})

		})

		grp.GET("/dashboard/plans", func(c *gin.Context) {

			address := validateSessionID(c)
			if address == "" {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			plans := db.GetAllSubscriptionPlans(_db)
			if plans == nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Failed to fetch subscription plans",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"plans": plans,
			})

		})

		grp.GET("/dashboard/plan", func(c *gin.Context) {

			address := validateSessionID(c)
			if address == "" {
				c.Redirect(http.StatusTemporaryRedirect, "/v1/login")
				return
			}

			if plan := db.CheckSubscriptionPlanDetailsByAddress(_db, common.HexToAddress(address)); plan != nil {
				c.JSON(http.StatusOK, plan)
				return
			}

			c.JSON(http.StatusNoContent, gin.H{
				"msg": "No Subscription plan found",
			})

		})

		// For checking `ette`'s syncing status
		grp.GET("/synced", func(c *gin.Context) {

			currentBlockNumber := db.GetCurrentBlockNumber(_db)
			blockCountInDB := _status.BlockCountInDB()
			remaining := (currentBlockNumber + 1) - blockCountInDB
			elapsed := _status.ElapsedTime()

			if cfg.Get("EtteMode") == "2" {
				c.JSON(http.StatusOK, gin.H{
					"processed": _status.Done(),
					"elapsed":   elapsed.String(),
				})
				return
			}

			status := fmt.Sprintf("%.2f %%", (float64(blockCountInDB)/float64(currentBlockNumber+1))*100)
			eta := "0s"
			if remaining > 0 {
				eta = (time.Duration((elapsed.Seconds()/float64(_status.Done()))*float64(remaining)) * time.Second).String()
			}

			c.JSON(http.StatusOK, gin.H{
				"synced":    status,
				"processed": _status.Done(),
				"elapsed":   elapsed.String(),
				"eta":       eta,
			})

		})

		// Query block data using block hash/ number/ block number range ( 10 at max )
		grp.GET("/block", checkEtteHistoricalMode, validateAPIKey, func(c *gin.Context) {

			hash := c.Query("hash")
			number := c.Query("number")
			tx := c.Query("tx")

			// Block hash based all tx retrieval request handler
			if strings.HasPrefix(hash, "0x") && len(hash) == 66 && tx == "yes" {
				if tx := db.GetTransactionsByBlockHash(_db, common.HexToHash(hash)); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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
					respondWithJSON(tx.ToJSON(), c)
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
					respondWithJSON(block.ToJSON(), c)
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
					respondWithJSON(block.ToJSON(), c)
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

				_from, _to, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if blocks := db.GetBlocksByNumberRange(_db, _from, _to); blocks != nil {
					respondWithJSON(blocks.ToJSON(), c)
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

				_from, _to, err := rangeChecker(fromTime, toTime, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if blocks := db.GetBlocksByTimeRange(_db, _from, _to); blocks != nil {
					respondWithJSON(blocks.ToJSON(), c)
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
		grp.GET("/transaction", checkEtteHistoricalMode, validateAPIKey, func(c *gin.Context) {

			hash := c.Query("hash")

			// Simply returns single tx object, when queried using tx hash
			if strings.HasPrefix(hash, "0x") && len(hash) == 66 {
				if tx := db.GetTransactionByHash(_db, common.HexToHash(hash)); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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
					respondWithJSON(tx.ToJSON(), c)
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

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetContractCreationTransactionsFromAccountByBlockNumberRange(_db, common.HexToAddress(deployer), _fromBlock, _toBlock); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetContractCreationTransactionsFromAccountByBlockTimeRange(_db, common.HexToAddress(deployer), _fromTime, _toTime); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetTransactionsBetweenAccountsByBlockNumberRange(_db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _fromBlock, _toBlock); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetTransactionsBetweenAccountsByBlockTimeRange(_db, common.HexToAddress(fromAccount), common.HexToAddress(toAccount), _fromTime, _toTime); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetTransactionsFromAccountByBlockNumberRange(_db, common.HexToAddress(fromAccount), _fromBlock, _toBlock); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetTransactionsFromAccountByBlockTimeRange(_db, common.HexToAddress(fromAccount), _fromTime, _toTime); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if tx := db.GetTransactionsToAccountByBlockNumberRange(_db, common.HexToAddress(toAccount), _fromBlock, _toBlock); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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

				_fromTime, _toTime, err := rangeChecker(fromBlock, toBlock, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if tx := db.GetTransactionsToAccountByBlockTimeRange(_db, common.HexToAddress(toAccount), _fromTime, _toTime); tx != nil {
					respondWithJSON(tx.ToJSON(), c)
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
		grp.GET("/event", checkEtteHistoricalMode, validateAPIKey, func(c *gin.Context) {

			fromBlock := c.Query("fromBlock")
			toBlock := c.Query("toBlock")

			fromTime := c.Query("fromTime")
			toTime := c.Query("toTime")

			contract := c.Query("contract")

			count := c.Query("count")

			topic0 := c.Query("topic0")
			topic1 := c.Query("topic1")
			topic2 := c.Query("topic2")
			topic3 := c.Query("topic3")

			blockHash := c.Query("blockHash")
			txHash := c.Query("txHash")

			// specific event log which under is requesting
			logIndex := c.Query("logIndex")

			// Given block hash and log index in block attempts to find out event
			// which satisfies that criteria
			if logIndex != "" && strings.HasPrefix(blockHash, "0x") && len(blockHash) == 66 {

				_logIndex, err := strconv.ParseUint(logIndex, 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad log index",
					})
					return
				}

				if event := db.GetEventByBlockHashAndLogIndex(_db, common.HexToHash(blockHash), uint(_logIndex)); event != nil {
					respondWithJSON(event.ToJSON(), c)
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return

			}

			// Given blockhash, retrieves all events emitted by tx present in block
			if strings.HasPrefix(blockHash, "0x") && len(blockHash) == 66 {

				if event := db.GetEventsByBlockHash(_db, common.HexToHash(blockHash)); event != nil {
					respondWithJSON(event.ToJSON(), c)
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return

			}

			// Given txhash, retrieves all events emitted by that tx ( i.e. during tx execution )
			if strings.HasPrefix(txHash, "0x") && len(txHash) == 66 {

				if event := db.GetEventsByTransactionHash(_db, common.HexToHash(txHash)); event != nil {
					respondWithJSON(event.ToJSON(), c)
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return

			}

			// Finds out last `x` events emitted by contract
			if count != "" && strings.HasPrefix(contract, "0x") && len(contract) == 42 {

				_count, err := strconv.Atoi(count)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad event count",
					})
					return
				}

				if _count > 50 {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Too many events requested",
					})
					return
				}

				if event := db.GetLastXEventsFromContract(_db, common.HexToAddress(contract), _count); event != nil {
					respondWithJSON(event.ToJSON(), c)
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return

			}

			// Given block number range, contract address & topics of log event, returns
			// events satisfying criteria
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(contract, "0x") && len(contract) == 42 && ((strings.HasPrefix(topic0, "0x") && len(topic0) == 66) || (strings.HasPrefix(topic1, "0x") && len(topic1) == 66) || (strings.HasPrefix(topic2, "0x") && len(topic2) == 66) || (strings.HasPrefix(topic3, "0x") && len(topic3) == 66)) {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if event := db.GetEventsFromContractWithTopicsByBlockNumberRange(_db, common.HexToAddress(contract), _fromBlock, _toBlock, getTopics([]string{topic0, topic1, topic2, topic3}...)...); event != nil {
					respondWithJSON(event.ToJSON(), c)
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return

			}

			// Given time span, contract address & topic 0 of log event, returns
			// events satisfying criteria
			if fromTime != "" && toTime != "" && strings.HasPrefix(contract, "0x") && len(contract) == 42 && ((strings.HasPrefix(topic0, "0x") && len(topic0) == 66) || (strings.HasPrefix(topic1, "0x") && len(topic1) == 66) || (strings.HasPrefix(topic2, "0x") && len(topic2) == 66) || (strings.HasPrefix(topic3, "0x") && len(topic3) == 66)) {

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if event := db.GetEventsFromContractWithTopicsByBlockTimeRange(_db, common.HexToAddress(contract), _fromTime, _toTime, getTopics([]string{topic0, topic1, topic2, topic3}...)...); event != nil {
					respondWithJSON(event.ToJSON(), c)
					return
				}

				c.JSON(http.StatusNotFound, gin.H{
					"msg": "Not found",
				})
				return

			}

			// Given block number range & contract address, finds out all events emitted by this contract
			if fromBlock != "" && toBlock != "" && strings.HasPrefix(contract, "0x") && len(contract) == 42 {

				_fromBlock, _toBlock, err := rangeChecker(fromBlock, toBlock, cfg.GetBlockNumberRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block number range",
					})
					return
				}

				if event := db.GetEventsFromContractByBlockNumberRange(_db, common.HexToAddress(contract), _fromBlock, _toBlock); event != nil {
					respondWithJSON(event.ToJSON(), c)
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

				_fromTime, _toTime, err := rangeChecker(fromTime, toTime, cfg.GetTimeRange())
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"msg": "Bad block time range",
					})
					return
				}

				if event := db.GetEventsFromContractByBlockTimeRange(_db, common.HexToAddress(contract), _fromTime, _toTime); event != nil {
					respondWithJSON(event.ToJSON(), c)
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

	upgrader := websocket.Upgrader{}

	router.GET("/v1/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[!] Failed to upgrade to websocket : %s\n", err.Error())
			return
		}

		// Registering websocket connection closing, to be executed when leaving
		// this function block
		defer conn.Close()

		if !checkEtteRealTimeMode(conn) {
			return
		}

		// Keeping track of which topics this client has subscribed to
		topics := make(map[string]ps.Consumer)
		lock := sync.Mutex{}

		// When returning from this execution scope, unsubscribing client
		// from all topics it might have subscribed to during it's life time
		//
		// Just attempting to do a graceful unsubscription
		defer func() {

			for _, v := range topics {

				if v, ok := v.(*ps.BlockConsumer); ok {
					v.Request.Type = "unsubscribe"
					continue
				}

				if v, ok := v.(*ps.TransactionConsumer); ok {
					v.Request.Type = "unsubscribe"
					continue
				}

				if v, ok := v.(*ps.EventConsumer); ok {
					v.Request.Type = "unsubscribe"
					continue
				}

			}

		}()

		// Client communication handling logic
		for {
			var req ps.SubscriptionRequest

			if err := conn.ReadJSON(&req); err != nil {
				log.Printf("[!] Failed to read message : %s\n", err.Error())
				break
			}

			// Validating client provided API key, if fails, we return
			// failure message to client & close connection
			user := req.GetUserFromAPIKey(_db)
			if user == nil {
				// -- Critical section of code begins
				//
				// Attempting to write to shared network connection
				lock.Lock()

				if err := conn.WriteJSON(&ps.SubscriptionResponse{Code: 0, Message: "Bad API Key"}); err != nil {
					log.Printf("[!] Failed to write message : %s\n", err.Error())
				}

				lock.Unlock()
				// -- ends here
				break
			}

			// Checking if user has kept this APIKey enabled or not
			if !user.Enabled {
				// -- Critical section of code begins
				//
				// Attempting to write to shared network connection
				lock.Lock()

				if err := conn.WriteJSON(&ps.SubscriptionResponse{Code: 0, Message: "Bad API Key"}); err != nil {
					log.Printf("[!] Failed to write message : %s\n", err.Error())
				}

				lock.Unlock()
				// -- ends here
				break
			}

			userAddress := common.HexToAddress(user.Address)

			// Checking if client is under allowed rate limit or not
			if !req.IsUnderRateLimit(_db, userAddress) {
				// -- Critical section of code begins
				//
				// Attempting to write to shared network connection
				lock.Lock()

				if err := conn.WriteJSON(&ps.SubscriptionResponse{Code: 0, Message: "Crossed Allowed Rate Limit"}); err != nil {
					log.Printf("[!] Failed to write message : %s\n", err.Error())
				}

				lock.Unlock()
				// -- ends here
				break
			}

			// Validating incoming request on websocket subscription channel
			if !req.Validate(topics) {
				// -- Critical section of code begins
				//
				// Attempting to write to shared network connection
				lock.Lock()

				if err := conn.WriteJSON(&ps.SubscriptionResponse{Code: 0, Message: "Bad Payload"}); err != nil {
					log.Printf("[!] Failed to write message : %s\n", err.Error())
				}

				lock.Unlock()
				// -- ends here
				break
			}

			switch req.Type {
			case "subscribe":
				switch req.Topic() {

				case "block":
					topics[req.Name] = ps.NewBlockConsumer(_redisClient, conn, &req, _db, userAddress, &lock)

				case "transaction":
					topics[req.Name] = ps.NewTransactionConsumer(_redisClient, conn, &req, _db, userAddress, &lock)

				case "event":
					topics[req.Name] = ps.NewEventConsumer(_redisClient, conn, &req, _db, userAddress, &lock)

				}
			case "unsubscribe":
				switch req.Topic() {

				case "block":
					if v, ok := topics[req.Name].(*ps.BlockConsumer); ok {
						v.Request.Type = req.Type
					}

				case "transaction":
					if v, ok := topics[req.Name].(*ps.TransactionConsumer); ok {
						v.Request.Type = req.Type
					}

				case "event":
					if v, ok := topics[req.Name].(*ps.EventConsumer); ok {
						v.Request.Type = req.Type
					}

				}

				delete(topics, req.Name)
			}
		}
	})

	router.POST("/v1/graphql", validateAPIKey, func(c *gin.Context) {

		gql := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
			Resolvers: &graph.Resolver{},
		}))

		if gql == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Failed to handle graphQL query",
			})
			return
		}

		gql.ServeHTTP(c.Writer, c.Request)

	})

	router.GET("/v1/graphql-playground", func(c *gin.Context) {

		if strings.ToLower(cfg.Get("EtteGraphQLPlayGround")) != "yes" {

			c.JSON(http.StatusOK, gin.H{
				"msg": "GraphQL Playground disabled",
			})
			return

		}

		gpg := playground.Handler("ette", "/v1/graphql")

		if gpg == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "Failed to create graphQL playground",
			})
			return
		}

		gpg.ServeHTTP(c.Writer, c.Request)

	})

	router.Run(fmt.Sprintf(":%s", cfg.Get("PORT")))
}
