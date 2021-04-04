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
	pb "github.com/itzmeanjan/ette/app/pb"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

// RestoreFromSnapshot - Given path to snapshot file and database handle
// where to restore entries, we're attempting to restore from snapshot file
//
// Reading from file is done sequentially, but processing each read byte array of
// our interest i.e. chunk which holds block data, to be processed concurrently
// using multiple workers
//
// Workers to be hired from worker pool of specific size.
func RestoreFromSnapshot(db *gorm.DB, file string) (bool, uint64) {

	// Opening file in read only mode
	fd, err := os.OpenFile(file, os.O_RDONLY, 0644)
	if err != nil {

		log.Printf("[!] Error : %s\n", err.Error())
		return false, 0

	}

	// file handle to be closed when whole file is read i.e.
	// EOF reached
	defer fd.Close()

	// count of entries read back from file
	var count uint64

	// creating worker pool, where each of them will attempt to
	// deserialize block data from byte array and put it into DB
	pool := wp.New(runtime.NumCPU() * int(cfg.GetConcurrencyFactor()))

	control := make(chan bool, 10000)
	entryCount := make(chan uint64)
	done := make(chan bool)

	go UnmarshalCoordinator(control, entryCount, done)

	for {

		buf := make([]byte, 4)

		// reading size of next protocol buffer encoded
		// data chunk
		if _, err := fd.Read(buf); err != nil {

			// reached EOF, good to get out of loop
			if err == io.EOF {
				break
			}

			log.Printf("[!] Failed to read chunk size : %s\n", err.Error())
			break

		}

		// converting size of next data chunk to `uint`
		// so that memory allocation can be performed
		// for next read
		size := binary.LittleEndian.Uint32(buf)

		// allocated buffer where to read next protocol buffer
		// serialized data chunk
		data := make([]byte, size)

		if _, err = fd.Read(data); err != nil {

			log.Printf("[!] Failed to read chunk : %s\n", err.Error())
			break

		}

		count++

		// Submitting job to worker pool
		// for unmarshalling data and putting structured form
		// into DB
		func(_data []byte) {

			pool.Submit(func() {
				ProcessBlock(db, _data, control)
			})

		}(data)

	}

	// letting coordinator know that `count` many workers
	// should let it know about their respective status of job
	entryCount <- count

	// waiting for coordinator to let us know
	// that all workers have completed their job
	<-done

	// no more jobs to be submitted to pool
	// but all existing one to be completed
	//
	// this call is redundant here, but still made
	pool.StopWait()

	return true, count

}

// ProcessBlock - Given byte array read from file, attempting
// to unmarshall it into structured data, which will be later
// attempted to be persisted into DB
//
// Also letting coordinator go routine know that this worker
// has completed its job
func ProcessBlock(db *gorm.DB, data []byte, control chan bool) {

	block := UnmarshalData(data)
	if block == nil {
		control <- false
		return
	}

	// attempting to create data struct of format which can be
	// easily used for persisting whole block data into DB
	_block := ProtoBufToBlock(block)

	if err := _db.StoreBlock(db, _block, nil, nil); err != nil {

		log.Printf("[!] Failed to restore block : %s\n", err.Error())

		control <- false
		return
	}

	control <- true

}

// UnmarshalData - Given byte array attempts to deserialize
// that into structured block data, which will be attempted to be
// written into DB
func UnmarshalData(data []byte) *pb.Block {

	block := &pb.Block{}
	if err := proto.Unmarshal(data, block); err != nil {

		log.Printf("[!] Failed to deserialize chunk : %s\n", err.Error())
		return nil

	}

	return block

}

// UnmarshalCoordinator - Given a lot of unmarshal workers to be
// created for processing i.e. deserialize & put into DB, more entries
// in smaller amount of time, they need to be synchronized properly
//
// That's all this go routine attempts to achieve
func UnmarshalCoordinator(control <-chan bool, count <-chan uint64, done chan bool) {

	// Letting main go routine know reading & processing all entries
	// done successfully, while getting out of this execution context
	defer func() {
		done <- true
	}()

	// received processing done count from worker go routines
	var success uint64
	// how many workers should actually let this go routine know
	// their status i.e. how long is this go routine supposed to
	// wait for all of them to finish
	var expected uint64

	for {
		select {

		case c := <-control:
			if !c {
				log.Fatalf("[!] Error received by unmarshal coordinator\n")
			}

			// some worker just let us know it completed
			// its job successfully
			success++
			// If this satisfies, it's time to exit from loop
			// i.e. all workers have completed their job
			if success == expected {
				return
			}

		// Once reading whole file is done, main go routine
		// knows how many entries are expected, which is to be
		// matched against how many of them actually completed their job
		//
		// Exiting from this loop, that logic is written 👆
		case v := <-count:
			expected = v
		}
	}

}
