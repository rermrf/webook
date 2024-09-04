token = nil

path = "/users/login"
method = "POST"

wrk.headers["Content-Type"] = "application/json"
wrk.headers["User-Agent"] = ""

request = function()
    body = '{"email": "123@qq.com", "password": "Hello@#world123"}'
    return wrk.format(method, path, wrk.headers, body)
end

response = function(status, headers, body)
    if not token and status == 200 then
        token = headers["X-Jwt-Token"]
        path = "/users/profile"
        method = "GET"
        wrk.headers["Authorization"]= string.format("Bear %s", token)
    end
end