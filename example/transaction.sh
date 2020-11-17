#!/bin/bash

# given tx hash, fetch tx details
curl http://localhost:3000/v1/transaction?hash=0x8a8f39c598c0ce929c8743fb883428eb1bf6f1afd9b189d958ca2cf5639d4e77

# for a specific account, given nonce, returns tx, in which nonce was set to
# this value for account given
curl http://localhost:3000/v1/transaction?nonce=1&fromAccount=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# can find out what contracts got deployed by specific account, when block number range is specified
#
# only 100 blocks you can ask it to query at max, at a given time
curl http://localhost:3000/v1/transaction?fromBlock=1&toBlock=100&deployer=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# can find out what contracts got deployed by specific account, when block number range is specified
#
# it can take 600 seconds of timespan at max, at a given time
curl http://localhost:3000/v1/transaction?fromTime=1604975929&toTime=1604975988&deployer=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d
