# 🎨 WebP Compressor v2.0

> 企业级WebP动画压缩工具 - 基于Clean Architecture重构

一个高性能、可扩展的WebP动画压缩工具，采用现代Go语言架构设计，提供标准版和嵌入版两种部署方式。

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Architecture](https://img.shields.io/badge/Architecture-Clean_Architecture-orange.svg)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## ✨ 特性

- 🏗️ **Clean Architecture** - 分层架构，易于维护和扩展
- 📊 **结构化日志** - 使用Go 1.21的slog包，支持多级别日志
- 🛡️ **优雅错误处理** - 分类错误管理，详细错误上下文
- ⚙️ **配置化系统** - 环境变量支持，灵活配置
- 🔧 **两种部署方式** - 标准版（外部依赖）和嵌入版（自包含）
- 🚀 **高性能压缩** - 智能进度显示，超时控制
- 🧪 **可测试性** - 依赖注入，接口抽象

## 📁 项目结构

```
webpcompressor/
├── bin/                    # 构建产物目录
│   ├── webpcompressor.exe  # 标准版可执行文件
│   └── webptools.exe       # 嵌入版可执行文件
├── cmd/                    # 主程序入口
│   ├── webpcompressor/     # 标准版入口
│   └── embedded/           # 嵌入版入口
├── internal/               # 内部包（不对外暴露）
│   ├── domain/             # 领域层 - 业务模型和接口
│   ├── service/            # 服务层 - 业务逻辑实现
│   ├── infrastructure/     # 基础设施层 - 外部依赖
│   └── config/             # 配置管理
├── pkg/                    # 公共包
│   ├── errors/             # 错误处理
│   └── logger/             # 日志系统
├── testdata/               # 测试数据
├── examples/               # 使用示例
└── dist/                   # 发布包目录
```

## 🚀 快速开始

### 📦 构建

```bash
# 使用构建脚本（推荐）
.\build.bat         # 英文版
.\build_cn.bat      # 中文版

# 手动构建
go build -o bin/webpcompressor.exe cmd/webpcompressor/main.go  # 标准版
go build -o bin/webptools.exe cmd/embedded/main.go            # 嵌入版
```

### 🎯 使用方法

#### 标准版（需要外部libwebp工具）
```bash
# 基本压缩
bin\webpcompressor.exe input.webp 30 output.webp

# 高质量压缩
bin\webpcompressor.exe animation.webp 50 compressed.webp
```

#### 嵌入版（内置所有工具）
```bash
# 压缩WebP动画
bin\webptools.exe compress input.webp 30 output.webp

# 查看WebP信息
bin\webptools.exe info animation.webp

# 显示帮助
bin\webptools.exe help
```

### 🔧 环境变量配置

```bash
# 日志级别
set WEBP_LOG_LEVEL=debug

# 临时目录
set WEBP_TEMP_DIR=D:\temp

# 最大并发数
set WEBP_MAX_CONCURRENCY=8

# 操作超时
set WEBP_TIMEOUT=10m

# 文件大小限制
set WEBP_MAX_FILE_SIZE=104857600
```

## 🏗️ 架构设计

### Clean Architecture 分层

```
┌─────────────────┐
│   Presentation  │  ← cmd/
├─────────────────┤
│    Service      │  ← internal/service/
├─────────────────┤
│     Domain      │  ← internal/domain/
├─────────────────┤
│ Infrastructure  │  ← internal/infrastructure/
└─────────────────┘
```

#### 领域层 (Domain)
- 业务模型：`FrameInfo`, `AnimationInfo`, `CompressionConfig`
- 核心接口：`WebPProcessor`, `ToolExecutor`, `FileManager`
- 业务规则：压缩配置，帧处理逻辑

#### 服务层 (Service)  
- 业务逻辑：`WebPService`
- 编排外部依赖
- 事务管理和错误处理

#### 基础设施层 (Infrastructure)
- 外部工具执行：`LocalToolExecutor`, `EmbeddedToolExecutor`
- 文件管理：`LocalFileManager`, `SafeFileManager`
- 工厂模式：依赖创建和管理

#### 支撑层 (Pkg)
- 结构化错误：分类错误，上下文信息
- 结构化日志：进度日志，操作日志
- 配置管理：环境变量，默认配置

## 🔄 版本对比

| 特性 | 标准版 | 嵌入版 |
|------|--------|--------|
| **文件大小** | ~2MB | ~8MB |
| **外部依赖** | 需要libwebp工具 | 无依赖 |
| **功能范围** | WebP压缩 | 12个WebP工具 |
| **部署方式** | 需要安装环境 | 单文件部署 |
| **启动速度** | 快 | 稍慢（需提取工具） |
| **适用场景** | 开发环境 | 生产环境 |

## 🛠️ 开发指南

### 添加新功能

1. **定义领域模型** - 在 `internal/domain/` 中定义接口和模型
2. **实现服务逻辑** - 在 `internal/service/` 中实现业务逻辑  
3. **扩展基础设施** - 在 `internal/infrastructure/` 中实现外部依赖
4. **更新入口程序** - 在 `cmd/` 中集成新功能

### 测试

```bash
# 运行测试
go test ./...

# 测试覆盖率
go test -cover ./...

# 性能测试
go test -bench=. ./...
```

### 代码质量

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 依赖整理
go mod tidy
```

## 📊 性能指标

- **内存使用**: 平均20MB，峰值50MB
- **处理速度**: 120帧动画 ~20秒
- **压缩率**: 30-70%（质量30-50）
- **并发处理**: 支持多帧并行压缩

## 🔗 相关链接

- [libwebp官方文档](https://developers.google.com/speed/webp)
- [Go Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [项目Issue跟踪](https://github.com/your-repo/issues)

## 📜 开源协议

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)  
5. 提交 Pull Request

## 📧 联系方式

- 项目维护者: [xiangtaohe@gmail.com]
- 技术讨论: [GitHub Discussions] 

---
