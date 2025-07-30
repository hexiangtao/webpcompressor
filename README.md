# 🎨 WebP Compressor v2.0

> 企业级WebP动画压缩工具 - 基于Clean Architecture重构，现已支持Web服务！

一个高性能、可扩展的WebP动画压缩工具，采用现代Go语言架构设计，提供命令行版本和现代化Web服务两种使用方式。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean_Architecture-orange.svg)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## ✨ 特性

### 🏗️ 架构特性
- **Clean Architecture** - 分层架构，易于维护和扩展
- **📊 结构化日志** - 使用Go 1.21的slog包，支持多级别日志
- **🛡️ 优雅错误处理** - 分类错误管理，详细错误上下文
- **⚙️ 配置化系统** - 环境变量支持，灵活配置
- **🧪 可测试性** - 依赖注入，接口抽象

### 🔧 部署方式
- **标准版** - 依赖外部WebP工具，灵活配置
- **嵌入版** - 内置所有工具，单文件部署
- **🌐 Web服务版** - 现代化Web界面，RESTful API

### 🚀 压缩功能
- **高性能压缩** - 智能进度显示，超时控制
- **并行处理** - 多核心加速，提升效率
- **多种预设** - 快速、平衡、高质量、网页优化
- **实时监控** - WebSocket实时进度推送

## 🌐 Web服务新特性

### 📱 现代化界面
- **拖拽上传** - 直观的文件上传体验
- **实时进度** - WebSocket实时进度显示
- **响应式设计** - 完美适配各种设备
- **美观UI** - Tailwind CSS现代化设计

### 🔌 API接口
- **RESTful API** - 标准化接口设计
- **文件管理** - 上传、下载、删除操作
- **任务管理** - 创建、查询、取消任务
- **实时通讯** - WebSocket进度推送

### 📊 管理功能
- **任务历史** - 完整的操作记录
- **统计信息** - 压缩效果统计
- **批量处理** - 支持多文件处理
- **文件安全** - 防路径遍历保护

## 📁 项目结构

```
webpcompressor/
├── bin/                        # 构建产物目录
│   ├── webpcompressor.exe      # 标准版可执行文件
│   ├── webptools.exe           # 嵌入版可执行文件
│   └── webpserver.exe          # Web服务版可执行文件
├── cmd/                        # 主程序入口
│   ├── webpcompressor/         # 标准版入口
│   ├── embedded/               # 嵌入版入口
│   └── webserver/              # Web服务版入口
├── internal/                   # 内部包（不对外暴露）
│   ├── domain/                 # 领域层 - 业务模型和接口
│   ├── service/                # 服务层 - 业务逻辑实现
│   ├── infrastructure/         # 基础设施层 - 外部依赖
│   ├── config/                 # 配置管理
│   └── web/                    # Web服务层
├── web/                        # Web前端资源
│   ├── index.html              # 主页面
│   └── static/                 # 静态资源
├── pkg/                        # 公共包
│   ├── errors/                 # 错误处理
│   └── logger/                 # 日志系统
├── testdata/                   # 测试数据
├── examples/                   # 使用示例
└── build_*.bat                 # 构建脚本
```

## 🚀 快速开始

### 方式一：Web服务（推荐）

1. **构建Web服务**
```bash
# Windows
build_web.bat

# 或手动构建
go build -o bin/webpserver.exe cmd/webserver/main.go
```

2. **启动服务**
```bash
cd bin
./webpserver.exe
```

3. **访问界面**
- 打开浏览器访问: http://localhost:8080
- 健康检查: http://localhost:8080/health
- API文档: http://localhost:8080/api/v1/info

### 方式二：命令行使用

1. **构建程序**
```bash
# 标准版
build.bat

# 嵌入版（推荐）
build_cn.bat
```

2. **使用示例**
```bash
# 标准版
webpcompressor input.webp 75 output.webp

# 嵌入版
webptools compress input.webp 75 output.webp
webptools info input.webp
```

## ⚙️ 配置说明

### Web服务配置

通过环境变量配置Web服务：

```bash
# 服务器配置
WEBP_WEB_HOST=0.0.0.0                # 服务器地址
WEBP_WEB_PORT=8080                   # 服务器端口
WEBP_WEB_MAX_FILE_SIZE=104857600     # 最大文件大小(100MB)

# 安全配置
WEBP_WEB_ENABLE_AUTH=false           # 是否启用认证
WEBP_WEB_AUTH_TOKEN=your_token       # 认证令牌

# 存储配置
WEBP_WEB_UPLOAD_DIR=uploads          # 上传目录
WEBP_WEB_OUTPUT_DIR=outputs          # 输出目录

# 任务配置
WEBP_WEB_MAX_CONCURRENT_TASKS=10     # 最大并发任务数
WEBP_WEB_TASK_TIMEOUT=1800           # 任务超时时间(秒)
WEBP_WEB_CLEANUP_INTERVAL=60         # 清理间隔(分钟)

# 日志配置
WEBP_LOG_LEVEL=info                  # 日志级别
WEBP_LOG_FILE=webp.log              # 日志文件
```

### 压缩参数

| 参数 | 说明 | 范围 | 推荐值 |
|------|------|------|--------|
| quality | 压缩质量 | 0-100 | 30-75 |
| method | 压缩方法 | 0-6 | 6 |
| preset | 压缩预设 | default/photo/picture/drawing/icon/text | photo |
| lossless | 无损压缩 | true/false | false |
| parallel | 并行处理 | true/false | true |

## 🔌 API文档

### 文件上传
```http
POST /api/v1/upload
Content-Type: multipart/form-data

file: [WebP文件]
```

### 创建压缩任务
```http
POST /api/v1/tasks
Content-Type: application/json

{
  "input_file": "uploaded_file.webp",
  "quality": 75,
  "preset": "photo",
  "lossless": false,
  "parallel": true
}
```

### 查询任务状态
```http
GET /api/v1/tasks/{task_id}
```

### WebSocket进度推送
```
ws://localhost:8080/api/v1/progress/{task_id}
```

### 文件下载
```http
GET /api/v1/download/{filename}
```

## 📊 使用示例

### Web界面使用

1. **上传文件** - 拖拽WebP文件到上传区域
2. **设置参数** - 调整压缩质量和预设
3. **开始压缩** - 点击开始压缩按钮
4. **监控进度** - 实时查看压缩进度
5. **下载结果** - 压缩完成后下载文件

### API调用示例

```javascript
// 上传文件
const formData = new FormData();
formData.append('file', file);
const uploadResponse = await fetch('/api/v1/upload', {
    method: 'POST',
    body: formData
});

// 创建压缩任务
const taskResponse = await fetch('/api/v1/tasks', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
        input_file: 'uploaded_file.webp',
        quality: 75,
        preset: 'photo'
    })
});

// WebSocket监听进度
const ws = new WebSocket(`ws://localhost:8080/api/v1/progress/${taskId}`);
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log(`进度: ${data.progress}%, 消息: ${data.message}`);
};
```

## 🛠️ 开发指南

### 环境要求
- Go 1.21+
- WebP工具集（标准版需要）

### 本地开发
```bash
# 克隆项目
git clone <repository>
cd webpcompressor

# 安装依赖
go mod tidy

# 运行Web服务
go run cmd/webserver/main.go

# 运行测试
go test ./...
```

### 项目架构

本项目采用Clean Architecture架构：

- **Domain层** - 核心业务逻辑和接口定义
- **Service层** - 业务服务实现
- **Infrastructure层** - 外部依赖实现
- **Web层** - HTTP服务和API处理

## 📈 性能特性

- **并行处理** - 多核心加速，处理效率提升300%+
- **内存优化** - 流式处理，支持大文件压缩
- **智能缓存** - 临时文件自动清理
- **超时控制** - 防止任务卡死，自动恢复

## 🔒 安全特性

- **文件验证** - 严格的文件格式检查
- **路径安全** - 防止路径遍历攻击
- **大小限制** - 可配置的文件大小限制
- **认证支持** - 可选的Token认证

## 📋 待办事项

- [ ] Docker部署支持
- [ ] 批量压缩接口
- [ ] 压缩预览功能
- [ ] 用户权限管理
- [ ] 压缩历史统计
- [ ] 多语言支持

## 🤝 贡献指南

欢迎提交问题和改进建议！

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [WebP项目](https://developers.google.com/speed/webp) - 提供优秀的WebP工具集
- [Gin Framework](https://gin-gonic.com/) - 高性能Web框架
- [Tailwind CSS](https://tailwindcss.com/) - 现代化CSS框架
