package data

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
)

// Event - Single event entity holder, extracted from db
type Event struct {
	Origin          string         `gorm:"column:origin"`
	Index           uint           `gorm:"column:index"`
	Topics          pq.StringArray `gorm:"column:topics;type:text[]"`
	Data            []byte         `gorm:"column:data"`
	TransactionHash string         `gorm:"column:txhash"`
	BlockHash       string         `gorm:"column:blockhash"`
}

// MarshalBinary - Implementing binary marshalling function, to be invoked
// by redis before publishing data on channel
func (e *Event) MarshalBinary() ([]byte, error) {
	return e.MarshalJSON()
}

// MarshalJSON - Custom JSON encoder
func (e *Event) MarshalJSON() ([]byte, error) {

	data := ""
	if _h := hex.EncodeToString(e.Data); _h != "" && _h != strings.Repeat("0", 64) {
		data = fmt.Sprintf("0x%s", _h)
	}

	return []byte(fmt.Sprintf(`{"origin":%q,"index":%d,"topics":%v,"data":%q,"txHash":%q,"blockHash":%q}`,
		e.Origin,
		e.Index,
		strings.Join(
			strings.Fields(
				fmt.Sprintf("%q", e.Topics)), ","),
		data, e.TransactionHash, e.BlockHash)), nil

}

// ToJSON - Encoding into JSON
func (e *Event) ToJSON() []byte {

	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("[!] Failed to encode event to JSON : %s\n", err.Error())
		return nil
	}

	return data

}

// Events - A collection of event holder, to be delivered to client in this form
type Events struct {
	Events []*Event `json:"events"`
}

// ToJSON - Encoding to JSON
func (e *Events) ToJSON() []byte {

	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("[!] Failed to encode events to JSON : %s\n", err.Error())
		return nil
	}

	return data

}
