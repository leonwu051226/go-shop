#!/bin/bash
echo "======================================"
echo "🚀 启动 MySQL TPS 实时监控 (订单落库速度)"
echo "======================================"

# 初始化统计
PREV_COUNT=$(docker exec seckill-mysql mysql -u seckill -pseckill123456 -D seckill -N -s -e "SELECT COUNT(*) FROM orders;")

while true; do
    sleep 1
    CURR_COUNT=$(docker exec seckill-mysql mysql -u seckill -pseckill123456 -D seckill -N -s -e "SELECT COUNT(*) FROM orders;")
    
    # 兼容处理空表的情况
    if [ -z "$CURR_COUNT" ]; then CURR_COUNT=0; fi
    if [ -z "$PREV_COUNT" ]; then PREV_COUNT=0; fi
    
    TPS=$((CURR_COUNT - PREV_COUNT))
    echo "$(date '+%H:%M:%S') | 当前总订单: $CURR_COUNT | 实时落库 TPS: $TPS 条/秒"
    PREV_COUNT=$CURR_COUNT
done
