package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	cfg "github.com/itzmeanjan/ette/app/config"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
)

func readConfig() error {
	configFile, err := filepath.Abs("../../.env")
	if err != nil {
		return errors.New(fmt.Sprintf("[!] Failed to find `.env` : %s\n", err.Error()))
	}

	err = cfg.Read(configFile)
	if err != nil {
		return errors.New(fmt.Sprintf("[!] Failed to read `.env` : %s\n", err.Error()))
	}
	return nil
}

func Test_NewClientPool(t *testing.T) {
	// TODO: must read config before getting clients
	err := readConfig()
	if err != nil {
		t.Fatal(err)
	}
	vitalikEth := "0xd8da6bf26964af9d7eed9e03e53415d37aa96045"

	clientPool := NewClientPool()
	client, err := clientPool.GetAvailableClient(true)
	if err != nil {
		t.Fatal(err)
	}
	balanceAt, err := client.GetClient().BalanceAt(context.Background(), common.HexToAddress(vitalikEth), nil)
	if err != nil {
		client.setError(err, time.Now().Add(1*time.Minute))
		t.Fatal(err)
	}
	t.Log(balanceAt.Int64())
}

func Test_SetError(t *testing.T) {
	// TODO: must read config before getting clients
	err := readConfig()
	if err != nil {
		t.Fatal(err)
	}
	clientPool := NewClientPool()
	client, err := clientPool.GetAvailableClient(true)
	if err != nil {
		t.Fatal(err)
	}
	client.setError(errors.New("reach limit exceed"), time.Now().Add(1*time.Minute))

	otherClient, err := clientPool.GetAvailableClient(true)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, client.rpcUrl, otherClient.rpcUrl)
}
