package data

// ActiveSubscriptions - Keeps track of how many active websocket
// connections being maintained now by `ette`
type ActiveSubscriptions struct {
	Count uint64
}
