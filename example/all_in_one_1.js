const { client } = require('websocket')

const _client = new client()

let state = true

_client.on('connectFailed', e => { console.error(`[!] Failed to connect : ${e}`); process.exit(1); })

// connect for listening to any block being mined
// event & any transaction being mined in any of those blocks
// & any event being emitted from contract interaction transactions
_client.on('connect', c => {
    c.on('close', d => {
        console.log(`[!] Closed connection : ${d}`)
        process.exit(0)
    })

    // receiving json encoded message
    c.on('message', d => { console.log(JSON.parse(d.utf8Data)) })

    // periodic subscription & unsubscription request performed
    handler = _ => {

        c.send(JSON.stringify({ name: 'block', type: state ? 'subscribe' : 'unsubscribe', apiKey: '0xab84c27c721f54ef3ea63f1f75ee70f1f0a2863ec30df5a9f1f057ccbba5d4c3' }))
        c.send(JSON.stringify({ name: 'transaction/*/*', type: state ? 'subscribe' : 'unsubscribe', apiKey: '0xab84c27c721f54ef3ea63f1f75ee70f1f0a2863ec30df5a9f1f057ccbba5d4c3' }))
        c.send(JSON.stringify({ name: 'event/*/*/*/*/*', type: state ? 'subscribe' : 'unsubscribe', apiKey: '0xab84c27c721f54ef3ea63f1f75ee70f1f0a2863ec30df5a9f1f057ccbba5d4c3' }))

        state = !state

    }

    setInterval(handler, 10000)
    handler()

})

_client.connect('ws://localhost:7000/v1/ws', null)
