# ette

Ethereum Blockchain Analyser 😎

## Table of Contents

- [Why did you build `ette` ?](#inspiration)
- [What do I need to have to use it ?](#prerequisite)
- [How to install it ?](#installation)
- [How to use it ?](#usage)

## Inspiration

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

## Prerequisite

- Make sure you've Go _( >= 1.15 )_ installed
- You need to also install & set up PostgreSQL. I found [this](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-20-04) guide helpful.
- Redis needs to be installed too. Consider following [this](https://www.digitalocean.com/community/tutorials/how-to-install-and-secure-redis-on-ubuntu-20-04) guide.
- Blockchain Node's websocket connection URL, required because we'll be listening for events in real time.

## Installation

- First fork this repository & clone it, some where out side of **GOPATH**.

```bash
git clone git@github.com:username/ette.git
```

- Now get insider `ette`

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

If everything goes as expected, you'll find one binary named, **ette** in this directory. Run it. 

```bash
./ette
```

> Note: For production, you'll most probably run it using `systemd`

## Usage

`ette` exposes REST API for querying historical block, transaction & event related data. It can also play role of real time notification engine, when subscribed to supported topics.

### Historical Block Data 🤩

You can query historical block data with various combination of query string params. 👇 is a comprehensive guide for consuming block data.

1) For fetching all transactions present in a block, when block hash is known

```bash
curl http://localhost:7000/v1/block?hash=0x...&tx=yes
```

2) For fetching all transactions present in a block, when block number is known

```bash
curl http://localhost:7000/v1/block?number=1&tx=yes
```

3) Fetching block by hash

```bash
curl http://localhost:7000/v1/block?hash=0x...
```

4) Fetching block by number

```bash
curl http://localhost:7000/v1/block?number=1
```

5) For fetching blocks by block number range _( max 10 at a time )_

```bash
curl http://localhost:7000/v1/block?fromBlock=1&toBlock=10
```

6) For fetching blocks by unix timestamp range _( max 60 seconds timespan )_

> Note: Timestamp matched with block mining timestamp

```bash
curl http://localhost:7000/v1/block?fromTime=1604975929&toTime=1604975988
```

**More coming soon**
