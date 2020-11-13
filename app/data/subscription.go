package data

import (
	"log"
	"regexp"
	"strings"
)

// SubscriptionRequest - Real time data subscription/ unsubscription request
// needs to be sent in this form, from client application
type SubscriptionRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetRegex - Returns regex to be used for validating subscription request
func (s *SubscriptionRequest) GetRegex() *regexp.Regexp {
	pattern, err := regexp.Compile("^(block|(transaction(/(0x[a-zA-Z0-9]{40}|\\*)(/(0x[a-zA-Z0-9]{40}|\\*))?)?))$")
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

	return ""
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
func (s *SubscriptionRequest) DoesMatchWithPublishedTransactionData(tx *Transaction) bool {

	// --- This closure function tries to match with to field of published tx data
	//
	// to field might not be present in some tx(s), where contract get deployed
	matchToFieldInTx := func(to string) bool {
		status := false

		switch to {
		// match with any `to` address
		case "":
			status = true
		// match with any `to` address
		case "*":
			status = true
		// match with specific `to` address
		default:
			if to == tx.To {
				status = true
			}
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
	case "":
		status = matchToFieldInTx(filters[1])
	// match with any `from` address
	case "*":
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

	if pattern.MatchString(s.Name) {

		matches := pattern.FindStringSubmatch(s.Name)
		if strings.HasPrefix(matches[0], "block") {
			return true
		} else if strings.HasPrefix(matches[0], "transaction") {
			return true
		}

	}

	return false
}

// Validate - Validates request from client for subscription/ unsubscription
func (s *SubscriptionRequest) Validate(topics map[string]bool) bool {

	// --- Closure definition
	// Given associative array & key in array, checks whether entry exists or not
	// If yes, also return entry's boolean value
	checkEntryInAssociativeArray := func(name string) bool {
		v, ok := topics[name]
		if !ok {
			return false
		}

		return v
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
