package client

import (
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	cfg "github.com/itzmeanjan/ette/app/config"
)

type Client struct {
	client            *ethclient.Client
	rpcUrl            string
	err               error
	availableAt       time.Time
	isWebsocketClient bool
}

func (c *Client) isAvailable() bool {
	if c.err != nil || time.Now().Before(c.availableAt) {
		return false
	}
	return true
}

func (c *Client) setError(err error, availableAt time.Time) {
	c.err = err
	c.availableAt = availableAt
}

func (c *Client) GetClient() *ethclient.Client {
	if c.client == nil {
		log.Fatal("❗️ No available client")
	}
	return c.client
}

// Connect to blockchain different nodes, either using HTTP or Websocket connection
// depending upon true/ false, passed to function, respectively
func getClients() (httpClients []*Client, webSocketClients []*Client) {
	initClientFunc := func(isRPC bool) []*Client {
		clients := make([]*Client, 0)
		var urls []string
		if isRPC {
			urls = strings.Split(cfg.Get("RPCUrls"), ",")
		} else {
			urls = strings.Split(cfg.Get("WebsocketUrls"), ",")
		}
		isWebsocketClient := !isRPC
		for _, rpcUrl := range urls {
			// If the URL is empty, ignore the invalid eth client
			if rpcUrl == "" {
				continue
			}
			client, err := ethclient.Dial(rpcUrl)
			if err != nil {
				log.Fatalf("[!] Failed to connect to blockchain : %s\n", err.Error())
			}
			clients = append(clients, &Client{
				client:            client,
				err:               nil,
				availableAt:       time.Now(),
				rpcUrl:            rpcUrl,
				isWebsocketClient: isWebsocketClient,
			})
		}
		return clients
	}
	httpClients = initClientFunc(true)
	webSocketClients = initClientFunc(false)
	return httpClients, webSocketClients
}
