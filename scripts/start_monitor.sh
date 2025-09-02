#!/bin/bash

echo "Starting HibiscusIM with Monitoring System..."
echo

# 设置环境变量
export APP_ENV=development
export DB_DRIVER=sqlite
export DSN=file:hibiscus_monitor.db?cache=shared
export ADDR=:8000
export LOG_LEVEL=info
export ENABLE_METRICS=true
export ENABLE_TRACING=true
export ENABLE_SQL_ANALYSIS=true
export ENABLE_SYSTEM_MONITOR=true

echo "Environment Variables:"
echo "  APP_ENV=$APP_ENV"
echo "  DB_DRIVER=$DB_DRIVER"
echo "  DSN=$DSN"
echo "  ADDR=$ADDR"
echo "  LOG_LEVEL=$LOG_LEVEL"
echo "  ENABLE_METRICS=$ENABLE_METRICS"
echo "  ENABLE_TRACING=$ENABLE_TRACING"
echo "  ENABLE_SQL_ANALYSIS=$ENABLE_SQL_ANALYSIS"
echo "  ENABLE_SYSTEM_MONITOR=$ENABLE_SYSTEM_MONITOR"
echo

# 启动服务器
echo "Starting server..."
go run cmd/server/main.go -mode=development
