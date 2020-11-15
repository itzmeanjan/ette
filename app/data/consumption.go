package data

// Consumer - Block, transaction & event consumers all are supposed to
// implement these methods, which are all, required to listen for data
// and send it to client applications
type Consumer interface {
	Subscribe()
	Listen()
	Send()
	SendData()
}
