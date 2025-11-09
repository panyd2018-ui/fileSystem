# 文件管理系统

一个基于 Go 开发的现代化文件管理系统，提供 Web 界面，支持文件上传、下载和删除功能。

## 功能特性

- ✅ **文件上传**: 支持拖拽上传和点击上传，无文件大小限制
- ✅ **文件下载**: 一键下载文件
- ✅ **文件删除**: 安全删除文件，带确认提示
- ✅ **文件列表**: 表格形式展示文件，显示图标、文件名、大小、修改时间等信息
- ✅ **响应式设计**: 适配桌面和移动设备
- ✅ **现代化 UI**: 美观的渐变设计和流畅的动画效果
- ✅ **文件图标**: 根据文件类型自动显示对应图标

## 技术栈

- **后端**: Go 1.21+, Gorilla Mux
- **前端**: HTML5, CSS3, JavaScript (原生)
- **架构**: RESTful API

## 快速开始

### 开发模式运行

```bash
# 1. 安装依赖
go mod download

# 2. 运行服务器
go run main.go

# 3. 访问应用
# 打开浏览器访问: http://localhost:8080
```

### 编译部署

**Windows:**
```bash
# 方式1: 直接编译
go build -o filesystem.exe .

# 方式2: 使用编译脚本
build.bat
```

**Linux/macOS:**
```bash
# 方式1: 直接编译
go build -o filesystem .

# 方式2: 使用编译脚本
chmod +x build.sh
./build.sh
```

详细部署说明请查看 [DEPLOY.md](DEPLOY.md)

## 项目结构

```
.
├── main.go              # 主程序入口
├── config.go            # 配置管理
├── models.go            # 数据模型
├── handlers.go          # HTTP 请求处理器
├── middleware.go        # 中间件
├── utils.go             # 工具函数
├── go.mod              # Go 模块文件
├── static/             # 静态文件目录
│   ├── index.html     # 前端页面
│   ├── style.css      # 样式文件
│   └── script.js      # JavaScript 文件
├── config.json         # 配置文件（首次运行自动创建）
└── README.md          # 项目说明
```

## API 接口

### 获取文件列表
```
GET /api/files
```

### 上传文件
```
POST /api/upload
Content-Type: multipart/form-data
Body: file (文件)
```

### 下载文件
```
GET /api/download/{filename}
```

### 删除文件
```
DELETE /api/delete/{filename}
```

## 配置说明

### 配置文件

程序使用 `config.json` 配置文件来设置存储目录和端口。首次运行时会自动创建配置文件。

**config.json 示例：**
```json
{
  "storage_dir": "D:\\MyFiles",
  "port": ":8080"
}
```

**配置项说明：**
- `storage_dir`: 文件存储目录
  - 如果为空或未设置，默认使用系统的下载目录（Downloads）
  - Windows: `C:\Users\用户名\Downloads`
  - macOS/Linux: `~/Downloads`
- `port`: 服务器端口（默认: `:8080`）

**修改配置：**
1. 直接编辑 `config.json` 文件
2. 修改后重启服务器即可生效

**注意：**
- Windows 路径使用双反斜杠 `\\` 或正斜杠 `/`
- 确保配置的目录有读写权限

## 安全特性

- 文件名验证，防止路径遍历攻击
- CORS 支持
- 错误处理和验证
- 流式文件处理，支持大文件上传

## 使用示例

1. **上传文件**: 
   - 点击上传区域或"选择文件"按钮
   - 或直接拖拽文件到上传区域

2. **下载文件**: 
   - 在文件列表中点击"下载"按钮

3. **删除文件**: 
   - 在文件列表中点击"删除"按钮
   - 确认删除操作

## 注意事项

- 确保有足够的磁盘空间存储上传的文件
- 建议在生产环境中添加身份验证和授权机制
- 支持上传任意大小的文件，无大小限制
- 大文件上传时会使用流式处理，不会占用过多内存

## 许可证

MIT License

