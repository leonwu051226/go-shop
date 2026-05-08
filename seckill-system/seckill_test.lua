-- seckill_test.lua
-- 替换为你刚才通过登录接口获取的真实 Token
local token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiaXNzIjoic2Vja2lsbC1zeXN0ZW0iLCJleHAiOjE3NzgzMDAwNzUsImlhdCI6MTc3ODIxMzY3NX0.HNUsqU8fw2cdx20mh4vsIWNchUY2gzXRACHvr_OCMRU"

wrk.method = "POST"
wrk.body   = '{"seckill_product_id": 1}'
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Authorization"] = "Bearer " .. token

-- 每次请求前动态生成，这里可以直接用静态配置，因为我们压测的是并发拦截能力
request = function()
    return wrk.format(nil, nil, nil, nil)
end