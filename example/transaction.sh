#!/bin/bash

# given tx hash, fetch tx details
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?hash=0x8a8f39c598c0ce929c8743fb883428eb1bf6f1afd9b189d958ca2cf5639d4e77

# for a specific account, given nonce, returns tx, in which nonce was set to
# this value for account given
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?nonce=1&fromAccount=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# can find out what contracts got deployed by specific account, when block number range is specified
#
# only 100 blocks you can ask it to query at max, at a given time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromBlock=1&toBlock=100&deployer=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# can find out what contracts got deployed by specific account, when block number range is specified
#
# it can take 600 seconds of timespan at max, at a given time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromTime=1604975929&toTime=1604975988&deployer=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# find out all tx in block number range, where `from` & `to` fields of tx
# are fixed
#
# at max 100 blocks can be queried at a time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromBlock=1&toBlock=100&fromAccount=0x0214c5C41F5249166AAb9a12213Ac0E2119D2cc8&toAccount=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# find out all tx in block number range, where `from` field of tx is fixed
#
# at max 100 blocks can be queried at a time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromBlock=1&toBlock=100&fromAccount=0x0214c5C41F5249166AAb9a12213Ac0E2119D2cc8

# find out all tx in block number range, where to` fields of tx is fixed
#
# at max 100 blocks can be queried at a time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromBlock=1&toBlock=100&toAccount=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# find out all tx in happened time span, where `from` & `to` field of tx are fixed
#
# at max 100 blocks can be queried at a time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromTime=1604975929&toTime=1604975988&fromAccount=0x0214c5C41F5249166AAb9a12213Ac0E2119D2cc8&toAccount=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d

# find out all tx in happened time span, where `from` field of tx is fixed
#
# at max 100 blocks can be queried at a time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromTime=1604975929&toTime=1604975988&fromAccount=0x0214c5C41F5249166AAb9a12213Ac0E2119D2cc8

# find out all tx in happened time span, where to` field of tx is fixed
#
# at max 100 blocks can be queried at a time
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/transaction?fromTime=1604975929&toTime=1604975988&toAccount=0x193919f422a262b1edfd8ca2776b8cfb000bdd3d
