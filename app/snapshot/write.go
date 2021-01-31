package snapshot

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"runtime"

	wp "github.com/gammazero/workerpool"
	cfg "github.com/itzmeanjan/ette/app/config"
	_db "github.com/itzmeanjan/ette/app/db"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// TakeSnapshot - Given sink file path & number of blocks to be read from database
// attempts to concurrently read whole blocks i.e. block header, transactions & events
// and serialize them into protocol buffer format, which are written into file with
// their respective size prepended in 4 bytes of reserved space.
//
// This kind of encoding mechanism helps us in encoding & decoding efficiently while
// gracefully using resources i.e. buffered processing, we get to snapshot very large datasets
// while consuming too much memory.
func TakeSnapshot(db *gorm.DB, file string, start uint64, end uint64, count uint64) bool {

	// checking given block number range correctness
	if !(start <= end) {

		log.Printf("[!] Ascending block number range required\n")
		return false

	}

	// Truncating/ opening for write/ creating data file, where to store protocol buffer encoded data
	fd, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[!] Failed to open snapshot file : %s\n", err.Error())
		return false
	}

	// to be invoked when returning from this function scope
	defer fd.Close()

	// creating worker pool, who will fetch data from DB, given block number
	// and serialize them into binary which will be passed to writer go routine
	// for writing to file
	pool := wp.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))

	// at max `count` many full block data can stay in buffer
	// before getting written into file
	data := make(chan []byte, count)
	// writer go routine wil let this coordinator know
	// it has written everything into file
	done := make(chan bool)

	// starting writer go routine
	go PutIntoSink(fd, count, data, done)

	// attempting to fetch X blocks ( max ) at a time, by range
	//
	// @note This can be improved
	var step uint64 = 10000

	// stepping through blocks present in DB
	// and attempting to process them by block number range
	for i := start; i <= end; i += step {

		// fetch block numbers, given range & attempt to process them concurrently
		blocks := _db.GetAllBlockNumbersInRange(db, i, i+step-1)
		if blocks == nil {
			continue
		}

		for _, b := range blocks {

			// taking block number to be processed by this job
			// inside function's local scope, so that a copy stays in closure
			func(num uint64) {

				pool.Submit(func() {

					_block := _db.GetBlockByNumber(db, num)
					if _block == nil {
						return
					}

					_protocolBufferedBlock, err := proto.Marshal(BlockToProtoBuf(_block, db))
					if err != nil {

						log.Printf("[!] Failed to serialize block : %s\n", err.Error())
						return

					}

					data <- _protocolBufferedBlock

				})

			}(b)

		}

	}

	pool.StopWait()
	// Blocking call i.e. waiting for writer go routine
	// to complete its job
	<-done

	return true

}

// PutIntoSink - Given open file handle and communication channels, waits for receiving
// new data to be written to snapshot file. Works until all data is received, once done processing
// lets coordinator go routine know it has successfully persisted all contents into file.
//
// This writer runs as an independent go routine, which simply writes data to file handle.
func PutIntoSink(fd io.Writer, count uint64, data chan []byte, done chan bool) {

	// Letting coordinator know writing to file has been completed
	// or some kind of error has occurred
	//
	// To be invoked when getting out of this execution scope
	defer func() {
		done <- true
	}()

	// How many data chunks received over channel
	//
	// To be compared against data chunks which were supposed
	// to be received, before deciding whether it's time to get out of
	// below loop or not
	var iter uint64

	for d := range data {

		// received new data which needs to be written to file
		iter++

		// store size of message ( in bytes ), in a byte array first
		// then that's to be written on file handle
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(len(d)))

		// first write size of proto message in 4 byte space
		if _, err := fd.Write(buf); err != nil {

			log.Printf("[!] Failed to write chunk size : %s\n", err.Error())
			break

		}

		// then write actual message
		if _, err := fd.Write(d); err != nil {

			log.Printf("[!] Failed to write chunk : %s\n", err.Error())
			break

		}

		// As soon as this condition is met,
		// we can safely get out of this loop
		// i.e. denoting all processing has been done
		if iter == count {
			break
		}

	}

}
