package app

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	cfg "github.com/itzmeanjan/ette/app/config"
	d "github.com/itzmeanjan/ette/app/data"
)

// Holds definition for all REST API(s) to be exposed
func runHTTPServer(_lock *sync.Mutex, _synced *d.SyncState) {
	router := gin.Default()

	grp := router.Group("/v1")

	{
		// For checking whether `ette` has synced upto blockchain latest state or not
		grp.GET("/synced", func(c *gin.Context) {

			_lock.Lock()
			defer _lock.Unlock()

			c.JSON(200, gin.H{
				"synced": _synced.Synced,
			})

		})
	}

	router.Run(fmt.Sprintf(":%s", cfg.Get("PORT")))
}
