@echo off
chcp 65001 >nul
echo ========================================
echo 文件管理系统 - 卸载自启动
echo ========================================
echo.

:: 检查管理员权限
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo [错误] 需要管理员权限！
    echo 请右键以管理员身份运行此脚本
    echo.
    pause
    exit /b 1
)

:: 检查任务是否存在
schtasks /query /tn "文件管理系统" >nul 2>&1
if %errorlevel% neq 0 (
    echo 未找到任务计划，可能已经删除
    echo.
    pause
    exit /b 0
)

echo 正在删除任务计划...
schtasks /delete /tn "文件管理系统" /f

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo 卸载成功！
    echo ========================================
    echo 开机自启动已移除
    echo ========================================
) else (
    echo.
    echo [错误] 删除任务失败！
)

echo.
pause

