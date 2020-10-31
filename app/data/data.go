package data

// SyncState - Whether `ette` is synced with blockchain or not
type SyncState struct {
	Synced bool
}

// Block - Block related info to be delivered to client in this format
type Block struct {
	Hash       string `json:"hash"`
	Number     uint64 `json:"number"`
	Time       uint64 `json:"time"`
	ParentHash string `json:"parentHash"`
	Difficulty string `json:"difficulty"`
	GasUsed    uint64 `json:"gasUsed"`
	GasLimit   uint64 `json:"gasLimit"`
	Nonce      uint64 `json:"nonce"`
}
