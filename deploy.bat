@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

:: ============================================================
:: WeBook 社区项目 - Windows Docker 快速部署脚本
:: ============================================================
:: 使用方法:
::   deploy.bat              # 完整部署
::   deploy.bat infra        # 仅部署基础设施
::   deploy.bat services     # 仅部署微服务
::   deploy.bat stop         # 停止所有服务
::   deploy.bat clean        # 清理所有容器和数据
::   deploy.bat logs [svc]   # 查看日志
::   deploy.bat status       # 查看服务状态
:: ============================================================

set "PROJECT_ROOT=%~dp0"
set "COMPOSE_FILE=%PROJECT_ROOT%docker-compose.full.yaml"

:: 检查 Docker
where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker 未安装，请先安装 Docker Desktop
    exit /b 1
)

:: 检查参数
if "%1"=="" goto :deploy_all
if "%1"=="infra" goto :deploy_infra
if "%1"=="services" goto :deploy_services
if "%1"=="stop" goto :stop_all
if "%1"=="clean" goto :clean_all
if "%1"=="logs" goto :view_logs
if "%1"=="status" goto :view_status
if "%1"=="build" goto :build_services
if "%1"=="help" goto :show_help
if "%1"=="--help" goto :show_help
if "%1"=="-h" goto :show_help
goto :deploy_all

:deploy_infra
echo [INFO] 正在部署基础设施服务...

:: 创建目录
if not exist "%PROJECT_ROOT%script\mysql" mkdir "%PROJECT_ROOT%script\mysql"

:: 复制初始化脚本
if exist "%PROJECT_ROOT%deploy\mysql\init.sql" (
    copy /Y "%PROJECT_ROOT%deploy\mysql\init.sql" "%PROJECT_ROOT%script\mysql\init.sql" >nul
)

:: 启动基础设施
docker compose -f "%COMPOSE_FILE%" up -d mysql redis etcd mongo kafka elasticsearch

echo [INFO] 等待基础设施启动...
timeout /t 30 /nobreak >nul

echo [SUCCESS] 基础设施部署完成
goto :eof

:build_services
echo [INFO] 正在构建服务镜像...

set "SERVICES=bff user article code comment follow interactive credit notification search tag feed ranking reward payment oauth2 sms account cronjob"

for %%s in (%SERVICES%) do (
    echo [INFO] 构建 %%s 服务...
    docker compose -f "%COMPOSE_FILE%" build %%s
)

echo [SUCCESS] 服务镜像构建完成
goto :eof

:deploy_services
echo [INFO] 正在部署微服务...

:: 核心服务
echo [INFO] 启动核心服务...
docker compose -f "%COMPOSE_FILE%" up -d user code sms oauth2
timeout /t 5 /nobreak >nul

:: 应用服务
echo [INFO] 启动应用服务...
docker compose -f "%COMPOSE_FILE%" up -d article interactive comment follow tag feed ranking
timeout /t 5 /nobreak >nul

:: 扩展服务
echo [INFO] 启动扩展服务...
docker compose -f "%COMPOSE_FILE%" up -d credit notification search reward payment account cronjob
timeout /t 3 /nobreak >nul

:: BFF 网关
echo [INFO] 启动 BFF 网关...
docker compose -f "%COMPOSE_FILE%" up -d bff

echo [SUCCESS] 微服务部署完成
goto :eof

:deploy_all
echo ============================================
echo 开始完整部署 WeBook 社区项目
echo ============================================

call :deploy_infra
call :build_services
call :deploy_services

echo.
echo ============================================
echo [SUCCESS] WeBook 部署完成！
echo ============================================
echo.
echo 服务访问地址：
echo   BFF API:      http://localhost:8081
echo   Prometheus:   http://localhost:8082/metrics
echo.
echo 基础设施：
echo   MySQL:        localhost:13306 (root/root)
echo   Redis:        localhost:6379
echo   Etcd:         localhost:12379
echo   Kafka:        localhost:9094
echo   Elasticsearch: localhost:9200
echo   MongoDB:      localhost:27017 (root/example)
echo.
echo 使用 'deploy.bat logs bff' 查看 BFF 服务日志
echo 使用 'deploy.bat status' 查看所有服务状态
goto :eof

:stop_all
echo [INFO] 正在停止所有服务...
docker compose -f "%COMPOSE_FILE%" down
echo [SUCCESS] 所有服务已停止
goto :eof

:clean_all
echo [WARN] 此操作将删除所有容器和数据卷，是否继续？(Y/N)
set /p confirm=
if /i "%confirm%"=="Y" (
    echo [INFO] 正在清理...
    docker compose -f "%COMPOSE_FILE%" down -v --rmi local
    echo [SUCCESS] 清理完成
) else (
    echo [INFO] 取消清理
)
goto :eof

:view_logs
if "%2"=="" (
    docker compose -f "%COMPOSE_FILE%" logs -f --tail=100
) else (
    docker compose -f "%COMPOSE_FILE%" logs -f --tail=100 %2
)
goto :eof

:view_status
echo [INFO] 服务状态：
docker compose -f "%COMPOSE_FILE%" ps
goto :eof

:show_help
echo WeBook Docker 部署脚本
echo.
echo 使用方法：
echo   deploy.bat              完整部署（基础设施 + 所有服务）
echo   deploy.bat infra        仅部署基础设施
echo   deploy.bat services     仅部署微服务
echo   deploy.bat build        仅构建服务镜像
echo   deploy.bat stop         停止所有服务
echo   deploy.bat clean        清理所有容器和数据
echo   deploy.bat logs [svc]   查看日志
echo   deploy.bat status       查看服务状态
echo   deploy.bat help         显示帮助信息
goto :eof
