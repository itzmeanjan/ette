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

// IsValidTopic - Checks whether topic to which client application is trying to
// subscribe to is valid one or not
func (s *SubscriptionRequest) IsValidTopic() bool {
	pattern, err := regexp.Compile("^(block|(transaction(/(0x[a-zA-Z0-9]{40}|\\*)(/(0x[a-zA-Z0-9]{40}|\\*))?)?))$")
	if err != nil {
		log.Printf("[!] Failed to parse regex pattern : %s\n", err.Error())
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
