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

		// Query block data using block hash
		grp.GET("/block", func(c *gin.Context) {

			hash := c.Query("hash")

			if hash != "" {
				if block := db.GetBlockByHash(_db, common.HexToHash(hash)); block != nil {

					data := block.ToJSON()
					if data == nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"msg": "JSON encoding failed",
						})
						return
					}

					c.Data(http.StatusOK, "application/json", data)
					return

				}

				c.JSON(http.StatusNoContent, gin.H{
					"msg": "Bad block hash",
				})
				return
			}

			number := c.Query("number")

			if number != "" {
				if block := db.GetBlockByNumber(_db, number); block != nil {

					data := block.ToJSON()
					if data == nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"msg": "JSON encoding failed",
						})
						return
					}

					c.Data(http.StatusOK, "application/json", data)
					return

				}

				c.JSON(http.StatusNoContent, gin.H{
					"msg": "Bad block hash",
				})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "Block hash/ number expected",
			})

		})
	}

	router.Run(fmt.Sprintf(":%s", cfg.Get("PORT")))
}
