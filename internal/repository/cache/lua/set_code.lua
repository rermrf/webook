-- phone_code:login:137XXXXXX
local key = KEYS[1]
-- 验证次数，最多重复三次，记录了还可以验证几次
-- phone_code:login:137XXXXXX:cnt
local cntKey = key..":cnt"
-- 验证码
local val = ARGV[1]

local ttl = tonumber(redis.call('ttl', key))

if ttl == -1 then
    -- key 存在，但是没有过期时间
    -- 系统错误，手动设置了这个 key，但是没有给过期时间
    return -2
    --    540 = 600 - 60 九分钟
elseif ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    -- 符合预期
    return 0
else
    -- 发送太频繁
    return -1
end

