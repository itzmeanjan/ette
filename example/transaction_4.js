const { client } = require('websocket')

const _client = new client()

let state = true

_client.on('connectFailed', e => { console.error(`[!] Failed to connect : ${e}`); process.exit(1); })

// For tx(s) happening from specific account to specific account
_client.on('connect', c => {
    c.on('close', d => {
        console.log(`[!] Closed connection : ${d}`)
        process.exit(0)
    })

    c.on('message', d => { console.log(JSON.parse(d.utf8Data)) })

    handler = _ => {

        // from & to address specified, any tx `from` -> `to` account can be
        // listened for using subscribing to this topic
        c.send(JSON.stringify({ name: 'transaction/0x4774fEd3f2838f504006BE53155cA9cbDDEe9f0c/0xc9D50e0a571aDd06C7D5f1452DcE2F523FB711a1', type: state ? 'subscribe' : 'unsubscribe', apiKey: '0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' }))
        state = !state

    }

    setInterval(handler, 10000)
    handler()

})

_client.connect('ws://localhost:7000/v1/ws', null)
