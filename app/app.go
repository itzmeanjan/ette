package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gookit/color"
	blk "github.com/itzmeanjan/ette/app/block"
	cfg "github.com/itzmeanjan/ette/app/config"
	"github.com/itzmeanjan/ette/app/db"

	"github.com/itzmeanjan/ette/app/rest"
	ss "github.com/itzmeanjan/ette/app/snapshot"
)

type Params struct {
	configFile, subscriptionPlansFile string
	down                              bool
}

// Run - Application to be invoked from main runner using this function
func Run(configFile, subscriptionPlansFile string) {

	ctx, cancel := context.WithCancel(context.Background())
	_connection, _redisClient, _redisInfo, _db, _status, _queue, _kafkaWriter := bootstrap(Params{configFile: configFile, subscriptionPlansFile: subscriptionPlansFile, down: false})

	// Attempting to listen to Ctrl+C signal
	// and when received gracefully shutting down `ette`
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGTERM, syscall.SIGINT)

	// RPC down channel
	rpc_info := make(chan bool)

	// All resources being used gets cleaned up
	// when we're returning from this function scope
	go func() {

		<-interruptChan

		// This call should be received in all places
		// where root context is passed along
		//
		// But only it's being used in block processor queue
		// go routine, as of now
		//
		// @note This can ( needs to ) be improved
		cancel()

		sql, err := _db.DB()
		if err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to get underlying DB connection : %s", err.Error()))
			return
		}

		if err := sql.Close(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to close underlying DB connection : %s", err.Error()))
			return
		}

		if err := _redisInfo.Client.Close(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to close connection to Redis : %s", err.Error()))
			return
		}

		if err := _kafkaWriter.Close(); err != nil {
			log.Print(color.Red.Sprintf("[!] Failed to close connection to kafka : %s", err.Error()))
			return
		}

		// Stopping process
		log.Print(color.Magenta.Sprintf("\n[+] Gracefully shut down `ette`"))
		os.Exit(0)

	}()

	// User has requested `ette` to take a snapshot of current database state
	if cfg.Get("EtteMode") == "4" {

		// checking if there's anything to snapshot or not
		if _status.BlockCountInDB() == 0 {
			log.Printf("[*] Nothing to snapshot\n")
			return
		}

		// this is the file snapshot to be taken
		_snapshotFile := cfg.GetSnapshotFile()
		_start := time.Now().UTC()

		log.Printf("[*] Starting snapshotting at : %s [ Sink : %s ]\n", _start, _snapshotFile)

		// taking snapshot, this might take some time
		_ret := ss.TakeSnapshot(_db, _snapshotFile, db.GetCurrentOldestBlockNumber(_db), db.GetCurrentBlockNumber(_db), _status.BlockCountInDB())
		if _ret {
			log.Print(color.Green.Sprintf("[+] Snapshotted in : %s [ Count : %d ]", time.Now().UTC().Sub(_start), _status.BlockCountInDB()))
		} else {
			log.Print(color.Red.Sprintf("[!] Snapshotting failed in : %s", time.Now().UTC().Sub(_start)))
		}

		return

	}

	// User has asked `ette` to attempt to restore from snapshotted data
	// where data file is `snapshot.bin` in current working directory,
	// if nothing specified for `SnapshotFile` variable in `.env`
	if cfg.Get("EtteMode") == "5" {

		_snapshotFile := cfg.GetSnapshotFile()
		_start := time.Now().UTC()

		log.Printf("[*] Starting snapshot restoring at : %s [ Sink : %s ]\n", _start, _snapshotFile)

		_, _count := ss.RestoreFromSnapshot(_db, _snapshotFile)

		log.Print(color.Green.Sprintf("[+] Restored from snapshot in : %s [ Count : %d ]", time.Now().UTC().Sub(_start), _count))

		return

	}

	go _queue.Start(ctx)

	// Pushing block header propagation listener to another thread of execution
	go blk.SubscribeToNewBlocks(_connection, _db, _status, _redisInfo, _queue, rpc_info)

	if <-rpc_info {
		_connection, _redisClient, _redisInfo, _db, _status, _queue, _kafkaWriter := bootstrap(Params{configFile: configFile, subscriptionPlansFile: subscriptionPlansFile, down: true})

		// User has requested `ette` to take a snapshot of current database state
		if cfg.Get("EtteMode") == "4" {

			// checking if there's anything to snapshot or not
			if _status.BlockCountInDB() == 0 {
				log.Printf("[*] Nothing to snapshot\n")
				return
			}

			// this is the file snapshot to be taken
			_snapshotFile := cfg.GetSnapshotFile()
			_start := time.Now().UTC()

			log.Printf("[*] Starting snapshotting at : %s [ Sink : %s ]\n", _start, _snapshotFile)

			// taking snapshot, this might take some time
			_ret := ss.TakeSnapshot(_db, _snapshotFile, db.GetCurrentOldestBlockNumber(_db), db.GetCurrentBlockNumber(_db), _status.BlockCountInDB())
			if _ret {
				log.Print(color.Green.Sprintf("[+] Snapshotted in : %s [ Count : %d ]", time.Now().UTC().Sub(_start), _status.BlockCountInDB()))
			} else {
				log.Print(color.Red.Sprintf("[!] Snapshotting failed in : %s", time.Now().UTC().Sub(_start)))
			}

			return

		}

		// User has asked `ette` to attempt to restore from snapshotted data
		// where data file is `snapshot.bin` in current working directory,
		// if nothing specified for `SnapshotFile` variable in `.env`
		if cfg.Get("EtteMode") == "5" {

			_snapshotFile := cfg.GetSnapshotFile()
			_start := time.Now().UTC()

			log.Printf("[*] Starting snapshot restoring at : %s [ Sink : %s ]\n", _start, _snapshotFile)

			_, _count := ss.RestoreFromSnapshot(_db, _snapshotFile)

			log.Print(color.Green.Sprintf("[+] Restored from snapshot in : %s [ Count : %d ]", time.Now().UTC().Sub(_start), _count))

			return

		}

		go blk.SubscribeToNewBlocks(_connection, _db, _status, _redisInfo, _queue, rpc_info)

		rest.RunHTTPServer(_db, _status, _redisClient, _kafkaWriter)

	}
	// Periodic clean up job being started, to be run every 24 hours to clean up
	// delivery history data, older than 24 hours
	//
	// @note Need to be diagnosed, why it doesn't work
	// go srv.DeliveryHistoryCleanUpService(_db)

	// Starting http server on main thread
	rest.RunHTTPServer(_db, _status, _redisClient, _kafkaWriter)

}
