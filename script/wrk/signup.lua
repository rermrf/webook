-- wrk -t1 -d1s -c2 -s ./script/wrk/signup.lua http://localhost:8081/users/signup
-- -t 线程数
-- -c 并发数
-- -d 持续时间，比如1s是一秒，1m也就是一分钟
-- -s 脚本路径

wrk.method="POST"
wrk.headers["Content-Type"] = "application/json"

local random = math.random
local function uuid()
    local template ='xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'
    return string.gsub(template, '[xy]', function (c)
        local v = (c == 'x') and random(0, 0xf) or random(8, 0xb)
        return string.format('%x', v)
    end)
end

-- 初始化
function init(args)
    -- 每个线程都有一个 cnt，所以是线程安全的
    cnt = 0
    prefix = uuid()
end

function request()
    body=string.format('{"email":"%s%d@qq.com", "password":"hello123.", "confirmPassword": "hello123."}', prefix, cnt)
    cnt = cnt + 1
    return wrk.format('POST', wrk.path, wrk.headers, body)
end

function response()

end