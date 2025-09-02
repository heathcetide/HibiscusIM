@echo off
echo Starting HibiscusIM with Monitoring System...
echo.

REM 设置环境变量
set APP_ENV=development
set DB_DRIVER=sqlite
set DSN=file:hibiscus_monitor.db?cache=shared
set ADDR=:8000
set LOG_LEVEL=info
set ENABLE_METRICS=true
set ENABLE_TRACING=true
set ENABLE_SQL_ANALYSIS=true
set ENABLE_SYSTEM_MONITOR=true

echo Environment Variables:
echo   APP_ENV=%APP_ENV%
echo   DB_DRIVER=%DB_DRIVER%
echo   DSN=%DSN%
echo   ADDR=%ADDR%
echo   LOG_LEVEL=%LOG_LEVEL%
echo   ENABLE_METRICS=%ENABLE_METRICS%
echo   ENABLE_TRACING=%ENABLE_TRACING%
echo   ENABLE_SQL_ANALYSIS=%ENABLE_SQL_ANALYSIS%
echo   ENABLE_SYSTEM_MONITOR=%ENABLE_SYSTEM_MONITOR%
echo.

REM 启动服务器
echo Starting server...
go run cmd/server/main.go -mode=development

pause
