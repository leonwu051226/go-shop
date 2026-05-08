-- seckill_test_stats.lua
local token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwiaXNzIjoic2Vja2lsbC1zeXN0ZW0iLCJleHAiOjE3NzgzMDI5MDgsImlhdCI6MTc3ODIxNjUwOH0.j7B2rVGC6TarRTStpn-zy6_y4H6bwB_D65wr02aVJPA"
wrk.method = "POST"
wrk.body   = '{"seckill_product_id": 1}'
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Authorization"] = "Bearer " .. token

-- 1. setup 阶段：给每个线程分配一个 ID 并记录，方便最后汇总
local threads = {}
function setup(thread)
    thread:set("id", #threads + 1)
    table.insert(threads, thread)
end

-- 2. init 阶段：初始化每个线程独立的计数器（避免多线程锁竞争）
function init(args)
    success_cnt = 0
    soldout_cnt = 0
    duplicate_cnt = 0
    other_error_cnt = 0
end

-- 3. response 阶段：每次收到 HTTP 响应时触发（极致性能的字符串匹配）
function response(status, headers, body)
    if status == 200 then
        -- 注意：这里的匹配字符串需要根据你 Go 后端真实的 JSON 返回值微调
        -- 假设成功返回 code: 0，库存不足返回 code: 400，已购买返回 403 等
        if string.find(body, '"code":0') or string.find(body, 'success') then
            success_cnt = success_cnt + 1
        elseif string.find(body, '库存不足') or string.find(body, 'sold out') then
            soldout_cnt = soldout_cnt + 1
        elseif string.find(body, '已购买') or string.find(body, '参与过') then
            duplicate_cnt = duplicate_cnt + 1
        else
            other_error_cnt = other_error_cnt + 1
        end
    else
        other_error_cnt = other_error_cnt + 1
    end
end

-- 4. done 阶段：压测结束时，汇总所有线程的数据并打印漂亮的结果
function done(summary, latency, requests)
    local total_success = 0
    local total_soldout = 0
    local total_duplicate = 0
    local total_other = 0

    -- 遍历所有线程，累加它们内部的计数器
    for _, thread in ipairs(threads) do
        total_success = total_success + thread:get("success_cnt")
        total_soldout = total_soldout + thread:get("soldout_cnt")
        total_duplicate = total_duplicate + thread:get("duplicate_cnt")
        total_other = total_other + thread:get("other_error_cnt")
    end

    io.write("\n==================================================\n")
    io.write("🎯 压测业务逻辑结果分析 (Business Logic Analysis)\n")
    io.write("==================================================\n")
    io.write(string.format("✅ 抢购成功 (排队中): %d\n", total_success))
    io.write(string.format("🛡️ 拦截: 已购买 (防重): %d\n", total_duplicate))
    io.write(string.format("🛡️ 拦截: 库存不足 (售罄): %d\n", total_soldout))
    io.write(string.format("❌ 其他异常 (HTTP 500等): %d\n", total_other))
    
    local total_processed = total_success + total_soldout + total_duplicate + total_other
    io.write("==================================================\n")
    io.write(string.format("总响应解析数: %d / %d (wrk 统计)\n", total_processed, summary.requests))
    io.write("==================================================\n")
end