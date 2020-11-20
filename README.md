# ette

Ethereum Blockchain Data Indexing Engine 😎

![banner](sc/banner.gif)

## Table of Contents

- [Why did you build `ette` ?](#inspiration-)
- [What do I need to have to use it ?](#prerequisite-)
- [How to install it ?](#installation-)
- [How to use it ?](#usage-)
    - Historical Data 
        - [Query historical block data](#historical-block-data-)
        - [Query historical transaction data](#historical-transaction-data-)
        - [Query historical event data](#historical-event-data-)
    - Real-time Data
        - [Real-time block mining notification](#real-time-notification-for-mined-blocks-)
        - [Real-time transaction notification ( 🤩 Filters Added ) ](#real-time-notification-for-transactions-%EF%B8%8F)
        - [Real-time log event notification ( 🤩 Filters Added ) ](#real-time-notification-for-events-)

## Inspiration 🤔

I was looking for one tool which will be able to keep itself in sync with latest happenings on EVM based blockchain, while exposing REST API for querying blockchain data with various filters. That tool will also expose real time notification functionalities over websocket, when subscribed to topics.

It's not that I was unable find any solution, but wasn't fully satisfied with those, so I decided to write `ette`, which will do following

- Sync upto latest state of blockchain
- Listen for all happenings on EVM based blockchain
- Persist all happenings in local database
- Expose REST API for querying 👇, while also setting block range/ time range for filtering results. Allow querying latest **X** entries for events emitted by contracts.
    - Block data
    - Transaction data
    - Event data

- Expose websocket based real time notification mechanism for 
    - Blocks being mined
    - Transactions being sent from address and/ or transactions being received at address
    - Events being emitted by contract, with indexed fields i.e. topics

And here's `ette`

## Prerequisite 👍

- Make sure you've Go _( >= 1.15 )_ installed
- You need to also install & set up PostgreSQL. I found [this](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-20-04) guide helpful.
- Redis needs to be installed too. Consider following [this](https://www.digitalocean.com/community/tutorials/how-to-install-and-secure-redis-on-ubuntu-20-04) guide.
- Blockchain Node's websocket connection URL, required because we'll be listening for events in real time.

## Installation 🛠

- First fork this repository & clone it, some where out side of **GOPATH**.

```bash
git clone git@github.com:username/ette.git
```

- Now get inside `ette`

```bash
cd ette
```

- Create a `.env` file in this directory.

```
RPC=wss://<websocket-endpoint>
PORT=7000
DB_USER=user
DB_PASSWORD=password
DB_HOST=x.x.x.x
DB_PORT=5432
DB_NAME=ette
RedisConnection=tcp
RedisAddress=x.x.x.x:6379
```

- Now build `ette`

```bash
go build
```

- If everything goes as expected, you'll find one binary named, **ette** in this directory. Run it. 

```bash
./ette
```

- Database migration to be taken care of during application start up.
- Syncing `ette` with latest state of blockchain takes time. Whether syncing is done or not, can be checked by querying

```bash
curl http://localhost:7000/v1/synced
```

> Note: For production, you'll most probably run it using `systemd`

## Usage 🦾

`ette` exposes REST API for querying historical block, transaction & event related data. It can also play role of real time notification engine, when subscribed to supported topics.

### Historical Block Data 🤩

You can query historical block data with various combination of query string params. 👇 is a comprehensive guide for consuming block data.

**Path : `/v1/block`**

**Example code snippet can be found [here](example/block.sh)**

Query Params | Method | Description
--- | --- | ---
`hash=0x...&tx=yes` | GET | Fetch all transactions present in a block, when block hash is known
`number=1&tx=yes` | GET | Fetch all transactions present in a block, when block number is known
`hash=0x...` | GET | Fetch block by hash
`number=1` | GET | Fetch block by number
`fromBlock=1&toBlock=10` | GET | Fetch blocks by block number range _( max 10 at a time )_
`fromTime=1604975929&toTime=1604975988` | GET | Fetch blocks by unix timestamp range _( max 60 seconds timespan )_

### Historical Transaction Data 😎

It's possible to query historical transactions data with various combination of query string params, where URL path is 👇

**Path : `/v1/transaction`**

**Example code snippet can be found [here](example/transaction.sh)**

Query Params | Method | Description
--- | --- | ---
`hash=0x...` | GET | Fetch transaction by txHash
`nonce=1&fromAccount=0x...` | GET | Fetch transaction, when tx sender's address & account nonce are known
`fromBlock=1&toBlock=10&deployer=0x...` | GET | Find out what contracts are created by certain account within given block number range _( max 100 blocks )_
`fromTime=1604975929&toTime=1604975988&deployer=0x...` | GET | Find out what contracts are created by certain account within given timestamp range _( max 600 seconds of timespan )_
`fromBlock=1&toBlock=100&fromAccount=0x...&toAccount=0x...` | GET | Given block number range _( max 100 at a time )_ & a pair of accounts, can find out all tx performed between that pair, where `from` & `to` fields are fixed
`fromTime=1604975929&toTime=1604975988&fromAccount=0x...&toAccount=0x...` | GET | Given time stamp range _( max 600 seconds of timespan )_ & a pair of accounts, can find out all tx performed between that pair, where `from` & `to` fields are fixed
`fromBlock=1&toBlock=100&fromAccount=0x...` | GET | Given block number range _( max 100 at a time )_ & an account, can find out all tx performed from that account
`fromTime=1604975929&toTime=1604975988&fromAccount=0x...` | GET | Given time stamp range _( max 600 seconds of span )_ & an account, can find out all tx performed from that account
`fromBlock=1&toBlock=100&toAccount=0x...` | GET | Given block number range _( max 100 at a time )_ & an account, can find out all tx where target was this address
`fromTime=1604975929&toTime=1604975988&toAccount=0x...` | GET | Given time stamp range _( max 600 seconds of span )_ & an account, can find out all tx where target was this address

### Historical Event Data 🧐

`ette` lets you query historical event data, emitted by smart contracts, by combination of query string params.

**Path : `/v1/event`**

Query Params | Method | Description
--- | --- | ---
`blockHash=0x...` | GET | Given blockhash, retrieves all events emitted by tx(s) present in block
`txHash=0x...` | GET | Given txhash, retrieves all events emitted during execution of this transaction
`count=50&contract=0x...` | GET | Returns last **x** _( <=50 )_ events emitted by this contract
`fromBlock=1&toBlock=10&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...&topic3=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0, 1, 2, 3}_
`fromBlock=1&toBlock=10&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0, 1, 2}_
`fromBlock=1&toBlock=10&contract=0x...&topic0=0x...&topic1=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0, 1}_
`fromBlock=1&toBlock=10&contract=0x...&topic0=0x...` | GET | Finding event(s) emitted from contract within given block range & also matching topic signatures _{0}_
`fromBlock=1&toBlock=10&contract=0x...` | GET | Finding event(s) emitted from contract within given block range
`fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...&topic3=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0, 1, 2, 3}_
`fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...&topic1=0x...&topic2=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0, 1, 2}_
`fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...&topic1=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0, 1}_
`fromTime=1604975929&toTime=1604975988&contract=0x...&topic0=0x...` | GET | Finding event(s) emitted from contract within given time stamp range & also matching topic signatures _{0}_
`fromTime=1604975929&toTime=1604975988&contract=0x...` | GET | Finding event(s) emitted from contract within given time stamp range

### Real time notification for mined blocks ⛏

For listening to blocks getting mined, connect to `/v1/ws` endpoint using websocket client library & once connected, you need to send **subscription** request with 👇 payload _( JSON encoded )_

```json
{
    "name": "block",
    "type": "subscribe"
}
```

If everything goes fine, your subscription will be confirmed with 👇 response _( JSON encoded )_

```json
{
    "code": 1,
    "message": "Subscribed to `block`"
}
```

After that as long as your machine is reachable, `ette` will keep notifying you about new blocks getting mined in 👇 form

```json
{
  "hash": "0x08f50b4795667528f6c0fdda31a0d270aae60dbe7bc4ea950ae1f71aaa01eabc",
  "number": 7015086,
  "time": 1605328635,
  "parentHash": "0x5ec0faff8b48e201e366a3f6c505eb274904e034c1565da2241f1327e9bad459",
  "difficulty": "6",
  "gasUsed": 78746,
  "gasLimit": 20000000,
  "nonce": 0
}
```

If you want to cancel subscription, consider sending 👇

```json
{
    "name": "block",
    "type": "unsubscribe"
}
```

You'll receive 👇 response, confirming unsubscription

```json
{
    "code": 1,
    "message": "Unsubscribed from `block`"
}
```

> Sample code can be found [here](example/block.js)

### Real time notification for transactions ⚡️

For listening to any transaction happening in network in real-time, send 👇 JSON encoded payload to `/v1/ws`

```json
{
    "name": "transaction/<from-address>/<to-address>",
    "type": "subscribe"
}
```

**Here we've some examples :**

- Any transaction

```json
{
    "name": "transaction/*/*",
    "type": "subscribe"
}
```

> Sample Code can be found [here](example/transaction_1.js)

- Fixed `from` field **[ tx originated `from` account ]**

```json
{
    "name": "transaction/0x4774fEd3f2838f504006BE53155cA9cbDDEe9f0c/*",
    "type": "subscribe"
}
```

> Sample Code can be found [here](example/transaction_2.js)

- Fixed `to` field **[ tx targeted `to` account ]**

```json
{
    "name": "transaction/*/0x4774fEd3f2838f504006BE53155cA9cbDDEe9f0c",
    "type": "subscribe"
}
```

> Sample Code can be found [here](example/transaction_3.js)

- Fixed `from` & `to` field **[ tx `from` -> `to` account ]**

```json
{
    "name": "transaction/0xc9D50e0a571aDd06C7D5f1452DcE2F523FB711a1/0x4774fEd3f2838f504006BE53155cA9cbDDEe9f0c",
    "type": "subscribe"
}
```

> Sample Code can be found [here](example/transaction_4.js)

If everything goes fine, your subscription will be confirmed with 👇 response _( JSON encoded )_

```json
{
    "code": 1,
    "message": "Subscribed to `transaction`"
}
```

After that as long as your machine is reachable, `ette` will keep notifying you about every transaction happening in 👇 form, where criterias matching

```json
{
  "hash": "0x08cfda79bd68ad280c7786e5dd349ab81981c52ea5cdd8e31be0a4b54b976555",
  "from": "0xc9D50e0a571aDd06C7D5f1452DcE2F523FB711a1",
  "to": "0x4774fEd3f2838f504006BE53155cA9cbDDEe9f0c",
  "contract": "",
  "gas": 200000,
  "gasPrice": "1000000000",
  "cost": "200000000000000",
  "nonce": 19899,
  "state": 1,
  "blockHash": "0xc29170d33141602a95b915c954c1068a380ef5169178eef2538beb6edb005810"
}
```

If you want to cancel subscription, consider sending 👇, while replacing `<from-address>` & `<to-address>` with specific addresses you used when subscribing.

```json
{
    "name": "transaction/<from-address>/<to-address>",
    "type": "unsubscribe"
}
```

You'll receive 👇 response, confirming unsubscription

```json
{
    "code": 1,
    "message": "Unsubscribed from `transaction`"
}
```

### Real-time notification for events 📧

For listening to any events getting emitted by smart contracts deployed on network, you need to send 👇 JSON encoded payload to `/v1/ws` endpoint, after connecting over websocket

```json
{
    "name": "event/<contract-address>/<topic-0-signature>/<topic-1-signature>/<topic-2-signature>/<topic-3-signature>",
    "type": "subscribe"
}
```

**Here we've some examples :**

- Any event emitted by any smart contract in network

```json
{
    "name": "event/*/*/*/*/*",
    "type": "subscribe"
}
```

- Any event emitted by one specific smart contract

```json
{
    "name": "event/0xcb3fA413B23b12E402Cfcd8FA120f983FB70d8E8/*/*/*/*",
    "type": "subscribe"
}
```

- Specific event emitted by one specific smart contract

```json
{
    "name": "event/0xcb3fA413B23b12E402Cfcd8FA120f983FB70d8E8/0x2ab93f65628379309f36cb125e90d7c902454a545c4f8b8cb0794af75c24b807/*/*/*",
    "type": "subscribe"
}
```

- Specific event emitted by any smart contract in network

```json
{
    "name": "event/*/0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef/*/*/*",
    "type": "subscribe"
}
```

> Sample code can be found [here](example/event_1.js)

If everything goes fine, your subscription will be confirmed with 👇 JSON encoded response

```json
{
    "code": 1,
    "message": "Subscribed to `event`"
}
```

After that as long as your machine is reachable, `ette` will keep notifying you about every event emitted by smart contracts, to which you've subscribed to, in 👇 format

```json
{
  "origin": "0x0000000000000000000000000000000000001010",
  "index": 3,
  "topics": [
    "0x4dfe1bbbcf077ddc3e01291eea2d5c70c2b422b415d95645b9adcfd678cb1d63",
    "0x0000000000000000000000000000000000000000000000000000000000001010",
    "0x0000000000000000000000004d31abd8533c00436b2145795cc4cef207c3364f",
    "0x00000000000000000000000042eefcda06ead475cde3731b8eb138e88cd0bac3"
  ],
  "data": "0x0000000000000000000000000000000000000000000000000000454b2247e2000000000000000000000000000000000000000000000000001a96ae0b49dfc60000000000000000000000000000000000000000000000003a0df005a45c3dd5dd0000000000000000000000000000000000000000000000001a9668c02797e40000000000000000000000000000000000000000000000003a0df04aef7e85b7dd",
  "txHash": "0xfdc5a29fdd57a53953a542f4c46b0ece5423227f26b1191e58d32973b4d81dc9",
  "blockHash": "0x08e9ac45e4041a4309c6f5dd42b0fc78e00ca0cb8603965465206b22a63d07fb"
}
```

If you want to cancel subscription, consider sending 👇, while replacing `<contract-address>`, `<topic-{0,1,2,3}-signature>` with specific values you used when subscribing.

```json
{
    "name": "event/<contract-address>/<topic-0-signature>/<topic-1-signature>/<topic-2-signature>/<topic-3-signature>",
    "type": "unsubscribe"
}
```

You'll receive 👇 response, confirming unsubscription

```json
{
    "code": 1,
    "message": "Unsubscribed from `event`"
}
```

> Note: If graceful unsubscription not done, when `ette` finds client unreachable, it'll remove client subscription

**More coming soon**
