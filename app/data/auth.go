package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	cfg "github.com/itzmeanjan/ette/app/config"
)

// AuthPayload - Payload to be sent in post request body, when performing
// login
type AuthPayload struct {
	Message   AuthPayloadMessage `json:"message" binding:"required"`
	Signature string             `json:"signature" binding:"required"`
}

// VerifySignature - Given recovered signer address from authentication payload
// checks whether person claiming to sign message has really signed or not
func (a *AuthPayload) VerifySignature(signer []byte) bool {
	if signer == nil {
		return false
	}

	return common.BytesToAddress(signer) == a.Message.Address
}

// IsAdmin - Given recovered signer address from authentication payload
// checks whether this address matches with admin address present in `.env` file
// or not
func (a *AuthPayload) IsAdmin(signer []byte) bool {
	if signer == nil {
		return false
	}

	return common.BytesToAddress(signer) == common.HexToAddress(cfg.Get("Admin"))
}

// HasExpired - Checking if message was signed with in
// `window` seconds ( will be kept generally 30s ) time span
// from current server time or not
func (a *AuthPayload) HasExpired(window int64) bool {
	return !(int64(a.Message.TimeStamp)+window >= time.Now().Unix())
}

// RecoverSigner - Given signed message & original message
// it recovers signer address as byte array, which is to be
// later used for matching against claimed signer address
func (a *AuthPayload) RecoverSigner() []byte {

	data := a.Message.ToJSON()
	if data == nil {
		return nil
	}

	signature, err := hexutil.Decode(a.Signature)
	if err != nil {
		return nil
	}

	if !(signature[64] == 27 || signature[64] == 28) {
		return nil
	}

	signature[64] -= 27

	pubKey, err := crypto.SigToPub(
		// After `Ethereum Signed Message` is prepended
		// we're performing keccak256 hash, which is actual message, signed in metamask
		crypto.Keccak256(
			// this is required, because for web3.personal.sign, it'll prepend this part
			// so we're also prepending it before recovering signature
			[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))),
		signature)
	if err != nil {
		return nil
	}

	return crypto.PubkeyToAddress(*pubKey).Bytes()

}

// AuthPayloadMessage - Message to be signed by user
type AuthPayloadMessage struct {
	Address   common.Address `json:"address" binding:"required"`
	TimeStamp uint64         `json:"timestamp" binding:"required"`
}

// ToJSON - Encoding message to JSON, this is what was signed by user
func (a *AuthPayloadMessage) ToJSON() []byte {

	if data, err := json.Marshal(a); err == nil {
		return data
	}

	return nil
}
