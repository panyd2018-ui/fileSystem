@echo off
chcp 65001 >nul
echo ========================================
echo 正在编译文件管理系统 (AMD64)...
echo ========================================
echo.

:: 设置编译环境变量
set GOOS=windows
set GOARCH=amd64

go build -o filesystem-amd64.exe .

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo 编译成功！
    echo ========================================
    echo 架构: AMD64 (64位)
    echo 操作系统: Windows
    echo 可执行文件: filesystem-amd64.exe
    echo.
    echo 部署文件清单:
    echo   - filesystem-amd64.exe (已包含前端文件)
    echo   - config.json (首次运行会自动创建)
    echo.
    echo 注意: 前端文件已打包进 exe，无需 static 目录
    echo ========================================
) else (
    echo.
    echo [错误] 编译失败！请检查错误信息。
    pause
    exit /b 1
)

