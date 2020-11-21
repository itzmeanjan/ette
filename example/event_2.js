const { client } = require('websocket')

const _client = new client()

let state = true

_client.on('connectFailed', e => { console.error(`[!] Failed to connect : ${e}`); process.exit(1); })

// For listening to any event being emitted
// by smart contract at `0xcb3fA413B23b12E402Cfcd8FA120f983FB70d8E8`
_client.on('connect', c => {
    c.on('close', d => {
        console.log(`[!] Closed connection : ${d}`)
        process.exit(0)
    })

    c.on('message', d => { console.log(JSON.parse(d.utf8Data)) })

    handler = _ => {

        c.send(JSON.stringify({ name: 'event/0xcb3fA413B23b12E402Cfcd8FA120f983FB70d8E8/*/*/*/*', type: state ? 'subscribe' : 'unsubscribe', apiKey: '0x43ee4523fc8cc569b492fcda6f32e07d90d2afdf7a8a02b5e74c933be1785487' }))
        state = !state

    }

    setInterval(handler, 10000)
    handler()

})

_client.connect('ws://localhost:7000/v1/ws', null)
