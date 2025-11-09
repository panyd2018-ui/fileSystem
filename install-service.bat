@echo off
chcp 65001 >nul
echo ========================================
echo 文件管理系统 - 开机自启动配置
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

:: 获取当前目录
set "CURRENT_DIR=%~dp0"
set "EXE_PATH=%CURRENT_DIR%filesystem.exe"

:: 检查文件是否存在
if not exist "%EXE_PATH%" (
    echo [错误] 找不到 filesystem.exe
    echo 请确保此脚本与 filesystem.exe 在同一目录
    echo.
    pause
    exit /b 1
)

echo 程序路径: %EXE_PATH%
echo 工作目录: %CURRENT_DIR%
echo.

:: 删除已存在的任务（如果存在）
schtasks /query /tn "文件管理系统" >nul 2>&1
if %errorlevel% equ 0 (
    echo 检测到已存在的任务，正在删除...
    schtasks /delete /tn "文件管理系统" /f >nul 2>&1
)

:: 创建任务计划
echo 正在创建任务计划...
schtasks /create /tn "文件管理系统" /tr "\"%EXE_PATH%\"" /sc onstart /ru "SYSTEM" /rl highest /f >nul 2>&1

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo 配置成功！
    echo ========================================
    echo 任务名称: 文件管理系统
    echo 启动方式: 系统启动时自动运行
    echo.
    echo 管理命令:
    echo   查看任务: schtasks /query /tn "文件管理系统"
    echo   删除任务: schtasks /delete /tn "文件管理系统" /f
    echo   立即运行: schtasks /run /tn "文件管理系统"
    echo ========================================
) else (
    echo.
    echo [错误] 创建任务失败！
    echo 请检查错误信息或手动配置
)

echo.
pause

