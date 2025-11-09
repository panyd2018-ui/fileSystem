#!/bin/bash

echo "========================================"
echo "正在编译文件管理系统 (Linux AMD64)..."
echo "========================================"
echo ""

# 设置编译环境变量
export GOOS=linux
export GOARCH=amd64

go build -o filesystem .

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "编译成功！"
    echo "========================================"
    echo "架构: AMD64 (64位)"
    echo "操作系统: Linux"
    echo "可执行文件: filesystem"
    echo ""
    echo "部署文件清单:"
    echo "  - filesystem (已包含前端文件)"
    echo "  - config.json (首次运行会自动创建)"
    echo ""
    echo "注意: 前端文件已打包进可执行文件，无需 static 目录"
    echo ""
    echo "运行方式: ./filesystem"
    echo "========================================"
    chmod +x filesystem
else
    echo ""
    echo "[错误] 编译失败！请检查错误信息。"
    exit 1
fi

