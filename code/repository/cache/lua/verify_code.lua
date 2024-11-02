local key = KEYS[1]
-- 用户输入的 code
local expectedCode = ARGV[1]
local code = redis.call("get", key)
local cntKey = key..":cnt"
local cnt = tonumber(redis.call("get", cntKey))
if cnt == nil or cnt <= 0 then
    -- 输错次数过多，直接pass
    return -1
elseif expectedCode == code then
    -- 输对了
    -- 用完不能再用
    redis.call("set", cntKey, -1)
    --redis.call("del", key)
    --redis.call("del", cntKey)
    return 0
else
    -- 输错了
    -- 可验证次数减1
    redis.call("decr", cntKey)
    return -2
end