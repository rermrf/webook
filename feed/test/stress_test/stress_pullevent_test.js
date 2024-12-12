import http from 'k6/http'

export let options = {
    duration: '10s',
    vus: 10,
    rpc: 10,
}

export default ()=>{
    let url = "http:127.0.0.1/feed/list"
    let payload = JSON.stringify({
        uid: 30001,
        limit: 10,
        timestamp: 1708748101,
    })

    let params = {
        headers: {
            'Content-Type': 'application/json',
        },
    }
    http.port(url, payload, params)
}