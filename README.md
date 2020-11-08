# ette
Ethereum Blockchain Analyser 😎

## Table of Contents

- [Why `ette` ?](#nomenclature)
- [Why did you build `ette` ?](#inspiration)

## Nomenclature

I was seaching for one catchy, single word name for my new project, which will expose some functionalities for analysing EVM based blockchain's historical data, will also provide real time notification for events happening on blockchain. And I came across the term `ETHER`, where I made a simple tweak & `ette` - I like the term because it reads same from both sides.

## Inspiration

I was looking for one tool which will be able to keep itself in sync with latest happenings on EVM based blockchain, while exposing REST API for querying blockchain data in several ways. That tool will also expose real time notification functionalities over websocket, when subscribed to different topics.

And I was able to find out some of my interest, but not in a single application. So I decided to write `ette`, which will do following things

- Listen for all happenings on EVM based blockchain
- Persist all happenings in local database
- Expose REST API for querying 👇, while also setting block range/ time range. Also allow querying latest **X** entries for events emitted by contracts. 
    - Block data
    - Transaction data
    - Event data

- Allow querying transaction data by sender/ receiver
- Expose websocket based real time notification mechanism for 
    - Blocks being mined
    - Transactions being sent from address and/ or transactions being received at address
    - Events being emitted by contract, with indexed fields i.e. topics

And here's `ette`

**More coming soon**
