@echo off
chcp 65001 >nul
echo ========================================
echo 文件管理系统 - 启动文件夹配置
echo ========================================
echo.

:: 获取当前目录和程序路径
set "CURRENT_DIR=%~dp0"
set "EXE_PATH=%CURRENT_DIR%filesystem.exe"
set "STARTUP_DIR=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup"

:: 检查文件是否存在
if not exist "%EXE_PATH%" (
    echo [错误] 找不到 filesystem.exe
    echo 请确保此脚本与 filesystem.exe 在同一目录
    echo.
    pause
    exit /b 1
)

echo 程序路径: %EXE_PATH%
echo 启动文件夹: %STARTUP_DIR%
echo.

:: 创建快捷方式
echo 正在创建快捷方式...
powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%STARTUP_DIR%\文件管理系统.lnk'); $s.TargetPath = '%EXE_PATH%'; $s.WorkingDirectory = '%CURRENT_DIR%'; $s.Save()"

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo 配置成功！
    echo ========================================
    echo 已添加到启动文件夹
    echo 程序将在用户登录时自动启动
    echo.
    echo 如需移除，删除以下文件：
    echo %STARTUP_DIR%\文件管理系统.lnk
    echo ========================================
) else (
    echo.
    echo [错误] 创建快捷方式失败！
)

echo.
pause

