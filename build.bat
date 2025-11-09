@echo off
echo 正在编译文件管理系统...
go build -o filesystem.exe .
if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo 编译成功！
    echo ========================================
    echo 可执行文件: filesystem.exe
    echo.
    echo 部署文件清单:
    echo   - filesystem.exe
    echo   - static\ 目录（包含所有前端文件）
    echo   - config.json（首次运行会自动创建）
    echo.
    echo 运行方式: filesystem.exe
    echo ========================================
) else (
    echo.
    echo 编译失败！请检查错误信息。
    pause
)

