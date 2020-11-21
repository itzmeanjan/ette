package pubsub

import (
	"log"
	"regexp"
	"strings"

	"github.com/itzmeanjan/ette/app/data"
)

// SubscriptionRequest - Real time data subscription/ unsubscription request
// needs to be sent in this form, from client application
type SubscriptionRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	APIKey string `json:"apiKey"`
}

// GetRegex - Returns regex to be used for validating subscription request
func (s *SubscriptionRequest) GetRegex() *regexp.Regexp {
	pattern, err := regexp.Compile("^(block|(transaction(/(0x[a-zA-Z0-9]{40}|\\*)(/(0x[a-zA-Z0-9]{40}|\\*))?)?)|(event(/(0x[a-zA-Z0-9]{40}|\\*)(/(0x[a-zA-Z0-9]{64}|\\*)(/(0x[a-zA-Z0-9]{64}|\\*)(/(0x[a-zA-Z0-9]{64}|\\*)(/(0x[a-zA-Z0-9]{64}|\\*))?)?)?)?)?))$")
	if err != nil {
		log.Printf("[!] Failed to parse regex pattern : %s\n", err.Error())
		return nil
	}

	return pattern
}

// Topic - Get main topic name to which this client is subscribing to
// i.e. {block, transaction}
func (s *SubscriptionRequest) Topic() string {
	if strings.HasPrefix(s.Name, "block") {
		return "block"
	}

	if strings.HasPrefix(s.Name, "transaction") {
		return "transaction"
	}

	if strings.HasPrefix(s.Name, "event") {
		return "event"
	}

	return ""
}

// GetLogEventFilters - Extracts contract address & topic signatures
// from subscription request, which are to be used
// for matching against published log event data
//
// Pattern looks like : `event/<address>/<topic0>/<topic1>/<topic2>/<topic3>`
//
// address : Contract address
// topic{0,1,2,3} : topic signature
func (s *SubscriptionRequest) GetLogEventFilters() []string {
	pattern := s.GetRegex()
	if pattern == nil {
		return nil
	}

	matches := pattern.FindStringSubmatch(s.Name)
	return []string{matches[9], matches[11], matches[13], matches[15], matches[17]}
}

// DoesMatchWithPublishedEventData  - All event channel listeners are going to get
// notified for any event emitted from smart contract interaction tx(s), but not all
// of clients are probably interested in those flood of data.
//
// Rather they've mentioned some filtering criterias, to be used for checking
// whether really this piece of data is requested by client or not
//
// All this function does, is checking whether it satisfies those criterias or not
func (s *SubscriptionRequest) DoesMatchWithPublishedEventData(event *data.Event) bool {

	// --- Matching specific topic signature provided by client
	// application with received event data, published by
	// redis pub-sub
	matchTopicXInEvent := func(topic string, x int) bool {
		// Not all topics will have 4 elements in topics array
		//
		// For those cases, if topic signature for that index is {"", "*"}
		// provided by consumer, then we're safely going to say it's a match
		if !(x < len(event.Topics)) {
			return topic == "" || topic == "*"
		}

		status := false

		switch topic {
		// match with any `topic` signature
		case "", "*":
			status = true
		// match with specific `topic` signature
		default:
			status = topic == event.Topics[x]
		}

		return status
	}
	// ---

	// Fetches desired filter values, against which matching to be performed
	// for published log event data
	filters := s.GetLogEventFilters()
	if filters == nil {
		return false
	}
	status := false

	switch filters[0] {
	// match with any `contract` address
	case "", "*":
		status = matchTopicXInEvent(filters[1], 0) && matchTopicXInEvent(filters[2], 1) && matchTopicXInEvent(filters[3], 2) && matchTopicXInEvent(filters[4], 3)
	// match with provided `contract` address
	default:
		if filters[0] == event.Origin {
			status = matchTopicXInEvent(filters[1], 0) && matchTopicXInEvent(filters[2], 1) && matchTopicXInEvent(filters[3], 2) && matchTopicXInEvent(filters[4], 3)
		}
	}

	return status

}

// GetTransactionFilters - Extracts from & to account present in transaction subscription request
//
// these could possibly be empty/ * / 0x...
func (s *SubscriptionRequest) GetTransactionFilters() []string {
	pattern := s.GetRegex()
	if pattern == nil {
		return nil
	}

	matches := pattern.FindStringSubmatch(s.Name)
	return []string{matches[4], matches[6]}
}

// DoesMatchWithPublishedTransactionData - All `transaction` topic listeners i.e. subscribers are
// going to get notified when new transaction detected, but they will only send those data to client application
// ( connected over websocket ), to which client has subscribed to
//
// Whether client has really shown interest in receiving notification for this transaction or not
// can be checked using this function
func (s *SubscriptionRequest) DoesMatchWithPublishedTransactionData(tx *data.Transaction) bool {

	// --- This closure function tries to match with to field of published tx data
	//
	// to field might not be present in some tx(s), where contract get deployed
	matchToFieldInTx := func(to string) bool {
		status := false

		switch to {
		// match with any `to` address
		case "", "*":
			status = true
		// match with specific `to` address
		default:
			status = to == tx.To
		}

		return status
	}
	// ---

	// Fetches desired to & from fields, against which matching to be performed
	filters := s.GetTransactionFilters()
	if filters == nil {
		return false
	}
	status := false

	switch filters[0] {
	// match with any `from` address
	case "", "*":
		status = matchToFieldInTx(filters[1])
	// match with provided `from` address
	default:
		if filters[0] == tx.From {
			status = matchToFieldInTx(filters[1])
		}
	}

	return status
}

// IsValidTopic - Checks whether topic to which client application is trying to
// subscribe to is valid one or not
func (s *SubscriptionRequest) IsValidTopic() bool {
	pattern := s.GetRegex()
	if pattern == nil {
		return false
	}

	return pattern.MatchString(s.Name)
}

// Validate - Validates request from client for subscription/ unsubscription
func (s *SubscriptionRequest) Validate(topics map[string]Consumer) bool {

	// --- Closure definition
	// Given associative array & key in array, checks whether entry exists or not
	// If yes, also return entry's boolean value
	checkEntryInAssociativeArray := func(topic string) bool {
		v, ok := topics[topic]
		if !ok {
			return false
		}

		return v != nil
	}
	// ---

	var validated bool

	switch s.Type {
	case "subscribe":
		validated = s.IsValidTopic() && !checkEntryInAssociativeArray(s.Name)
	case "unsubscribe":
		validated = s.IsValidTopic() && checkEntryInAssociativeArray(s.Name)
	default:
		validated = false
	}

	return validated

}

// SubscriptionResponse - Real time data subscription/ unsubscription request to be responded with
// in this form
type SubscriptionResponse struct {
	Code    uint   `json:"code"`
	Message string `json:"msg"`
}
