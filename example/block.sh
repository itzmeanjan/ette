#!/bin/bash

# fetches all tx mined in block, given block hash
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/block?hash=0x38db3fe4118acea4d61740ad80fd66018f248f9dc44374575aba3999dbbc0b25&tx=yes

# fetches all tx mined in block, given block number
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/block?number=7155907&tx=yes

# fetches only block related data, given block hash
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/block?hash=0x38db3fe4118acea4d61740ad80fd66018f248f9dc44374575aba3999dbbc0b25

# fetches only block related data, given block number
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/block?number=7155907

# Batch block fetching, by block number range/ timestamp range

# How many block can be queried at a time, is limited
# as of now, it's set to 10
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/block?fromBlock=1&toBlock=10

# How many block can be queried at a time, is limited
# as of now, set to 60 seconds timespan
curl -H 'APIKey: 0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' http://localhost:7000/v1/block?fromTime=1604975929&toTime=1604975988
