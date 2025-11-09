# 部署说明

## 编译

### Windows 编译

```bash
# 编译为 Windows AMD64 可执行文件（64位）
set GOOS=windows
set GOARCH=amd64
go build -o filesystem.exe .

# 或者直接使用编译脚本
build.bat
```

**注意：** 默认编译为当前系统架构。如需指定架构，请设置 `GOOS` 和 `GOARCH` 环境变量。

### Linux 编译

**在 Linux 系统上编译：**
```bash
# 使用编译脚本（自动设置为 AMD64）
chmod +x build.sh
./build.sh

# 或手动编译
export GOOS=linux
export GOARCH=amd64
go build -o filesystem .
```

**在 Windows 上交叉编译 Linux 版本：**
```bash
# 使用编译脚本
build-linux-amd64.bat

# 或手动编译
set GOOS=linux
set GOARCH=amd64
go build -o filesystem-linux-amd64 .
```

### macOS 编译

```bash
# 编译为 macOS 可执行文件
go build -o filesystem main.go

# 或者交叉编译（在 Windows 上编译 macOS 版本）
set GOOS=darwin
set GOARCH=amd64
go build -o filesystem main.go
```

## 部署文件清单

部署时需要包含以下文件：

```
filesystem.exe (或 filesystem)    # 编译后的可执行文件（已包含前端文件）
config.json                       # 配置文件（首次运行会自动创建）
```

**注意：** 前端文件已打包进可执行文件中，无需单独的 `static/` 目录。

## 部署步骤

### 1. 编译程序

```bash
go build -o filesystem.exe .
```

### 2. 准备部署目录

创建部署目录，将以下文件复制到目录中：
- 编译后的可执行文件（已包含前端文件）
- `config.json`（可选，首次运行会自动创建）

**注意：** 前端文件已嵌入到可执行文件中，无需单独部署 `static/` 目录。

### 3. 配置

编辑 `config.json` 设置存储目录和端口：

```json
{
  "storage_dir": "D:\\Files",
  "port": ":8080"
}
```

### 4. 运行

**Windows:**
```cmd
filesystem.exe
```

**Linux/macOS:**
```bash
chmod +x filesystem
./filesystem
```

### 5. 后台运行（Linux）

使用 `nohup` 或 `systemd` 服务：

```bash
# 使用 nohup
nohup ./filesystem > filesystem.log 2>&1 &

# 或使用 screen
screen -S filesystem
./filesystem
# 按 Ctrl+A 然后 D 退出 screen
```

### 6. 创建系统服务（Linux systemd）

创建服务文件 `/etc/systemd/system/filesystem.service`:

```ini
[Unit]
Description=File System Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/filesystem
ExecStart=/path/to/filesystem/filesystem
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable filesystem
sudo systemctl start filesystem
```

查看状态：
```bash
sudo systemctl status filesystem
```

## Windows 开机自启动

### 方法一：使用任务计划程序（推荐）

运行安装脚本：
```cmd
# 以管理员身份运行
install-service.bat
```

### 方法二：使用启动文件夹（简单）

运行安装脚本：
```cmd
install-startup.bat
```

### 卸载自启动

```cmd
# 卸载任务计划
uninstall-service.bat

# 或手动删除启动文件夹中的快捷方式
```

详细说明请查看 [WINDOWS_AUTOSTART.md](WINDOWS_AUTOSTART.md)

## 生产环境建议

1. **使用反向代理**（Nginx/Apache）
   - 配置 HTTPS
   - 设置域名
   - 负载均衡（如需要）

2. **防火墙配置**
   - 开放配置的端口（默认 8080）
   - 限制访问来源（如需要）

3. **权限设置**
   - 确保程序有读写存储目录的权限
   - 不要使用 root 用户运行（Linux）

4. **日志管理**
   - 配置日志轮转
   - 监控程序运行状态

5. **安全加固**
   - 添加身份验证（建议）
   - 限制文件大小（如需要）
   - 配置 CORS 白名单

## 快速部署脚本（Windows）

创建 `build.bat`:

```batch
@echo off
echo 正在编译...
go build -o filesystem.exe .
if %errorlevel% equ 0 (
    echo 编译成功！
    echo 可执行文件: filesystem.exe
) else (
    echo 编译失败！
    pause
)
```

## 快速部署脚本（Linux）

创建 `build.sh`:

```bash
#!/bin/bash
echo "正在编译..."
go build -o filesystem .
if [ $? -eq 0 ]; then
    echo "编译成功！"
    echo "可执行文件: filesystem"
    chmod +x filesystem
else
    echo "编译失败！"
    exit 1
fi
```

