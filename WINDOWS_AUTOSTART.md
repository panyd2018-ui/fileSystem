# Windows 开机自启动配置

## 方法一：任务计划程序（推荐）

这是最可靠的方法，可以设置程序在系统启动时自动运行。

### 手动配置步骤

1. **打开任务计划程序**
   - 按 `Win + R`，输入 `taskschd.msc`，回车

2. **创建基本任务**
   - 点击右侧"创建基本任务"
   - 名称：`文件管理系统`
   - 描述：`工程组文件管理系统自动启动`

3. **设置触发器**
   - 选择"当计算机启动时"

4. **设置操作**
   - 选择"启动程序"
   - 程序或脚本：浏览选择 `filesystem.exe` 的完整路径
   - 起始于：选择 `filesystem.exe` 所在的目录路径

5. **完成设置**
   - 勾选"当单击完成时，打开此任务属性的对话框"
   - 点击"完成"

6. **高级设置**（可选）
   - 在属性对话框中：
     - 勾选"使用最高权限运行"（如果需要）
     - 在"条件"选项卡中，取消勾选"只有在计算机使用交流电源时才启动此任务"（如果需要电池供电时也运行）
     - 在"设置"选项卡中，可以设置失败后重启任务

### 使用脚本自动配置

运行 `install-service.bat` 脚本可以自动创建任务计划。

## 方法二：启动文件夹（简单但不推荐）

### 步骤

1. **打开启动文件夹**
   - 按 `Win + R`，输入 `shell:startup`，回车

2. **创建快捷方式**
   - 在启动文件夹中，右键 → 新建 → 快捷方式
   - 浏览选择 `filesystem.exe`
   - 点击"下一步"，输入名称，点击"完成"

### 缺点

- 需要用户登录后才能启动
- 如果程序崩溃，不会自动重启
- 没有日志记录

## 方法三：使用 NSSM（Windows 服务）

NSSM 可以将普通程序安装为 Windows 服务，更专业。

### 安装步骤

1. **下载 NSSM**
   - 访问：https://nssm.cc/download
   - 下载适合的版本（32位或64位）

2. **解压 NSSM**
   - 解压到某个目录，如 `C:\nssm`

3. **安装服务**
   ```cmd
   # 以管理员身份运行 CMD
   cd C:\nssm\win64
   nssm install FileSystemService
   ```
   
   在弹出的界面中设置：
   - Path: `C:\path\to\filesystem.exe`（你的程序路径）
   - Startup directory: `C:\path\to\`（程序所在目录）
   - Service name: `FileSystemService`

4. **启动服务**
   ```cmd
   nssm start FileSystemService
   ```

5. **设置服务为自动启动**
   ```cmd
   sc config FileSystemService start= auto
   ```

### 管理服务

```cmd
# 启动服务
nssm start FileSystemService

# 停止服务
nssm stop FileSystemService

# 重启服务
nssm restart FileSystemService

# 删除服务
nssm remove FileSystemService confirm
```

## 方法四：使用 PowerShell 脚本

创建一个 PowerShell 脚本来自动配置任务计划。

## 推荐方案

- **个人使用/测试**：使用方法二（启动文件夹），简单快速
- **生产环境**：使用方法一（任务计划程序）或方法三（NSSM 服务），更稳定可靠

