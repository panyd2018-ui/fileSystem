@echo off
chcp 65001 >nul
echo ========================================
echo 正在编译文件管理系统 (Linux AMD64)...
echo ========================================
echo.

:: 设置编译环境变量
set GOOS=linux
set GOARCH=amd64

go build -o filesystem-linux-amd64 .

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo 编译成功！
    echo ========================================
    echo 架构: AMD64 (64位)
    echo 操作系统: Linux
    echo 可执行文件: filesystem-linux-amd64
    echo.
    echo 部署文件清单:
    echo   - filesystem-linux-amd64 (已包含前端文件)
    echo   - config.json (首次运行会自动创建)
    echo.
    echo 注意: 前端文件已打包进可执行文件，无需 static 目录
    echo.
    echo 部署到 Linux 后:
    echo   chmod +x filesystem-linux-amd64
    echo   ./filesystem-linux-amd64
    echo ========================================
) else (
    echo.
    echo [错误] 编译失败！请检查错误信息。
    pause
    exit /b 1
)

