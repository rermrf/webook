import http from 'k6/http';

export let options = {
    duration: '10s',
    vus: 100,
    rpc: 20,
}

export default () => {
    let url = "http://127.0.0.1:8088/feed/add";
    let jsonStr = '{"uid": "6", "aid": "article456", "title": "Example Title"}'
    let payload = JSON.stringify({
        typ: "article_event",
        ext: jsonStr
    })

    let params = {
        headers: {
            'Content-Type': 'application/json',
        }
    };
    http.post(url, payload, params)
}