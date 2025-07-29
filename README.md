# 🎨 WebP多功能工具集

一个强大的WebP动画压缩和处理工具集，提供两个版本：标准版和嵌入版。

## 📦 版本说明

### 🔧 标准版 (webpcompressor.exe)
- 依赖外部libwebp工具
- 专注WebP动画压缩
- 文件大小小
- 需要安装libwebp

### 🌟 嵌入版 (webptools.exe) 
- **内置所有工具**，无需外部依赖
- 支持12个WebP处理功能
- 一个文件解决所有需求
- 文件大小较大（约5MB）

## 🚀 快速开始

### 编译程序

```bash
# 使用构建脚本（推荐）
.\build.bat

# 或手动编译
# 标准版
go build -o webpcompressor.exe main.go

# 嵌入版  
go build -o webptools.exe cmd/embedded/main.go
```

### 使用示例

#### 标准版使用
```bash
# 压缩WebP动画
.\webpcompressor.exe input.webp 30 output.webp

# 查看帮助
.\webpcompressor.exe
```

#### 嵌入版使用
```bash
# 压缩WebP动画
.\webptools.exe compress input.webp 30 output.webp

# GIF转WebP
.\webptools.exe convert animation.gif animation.webp 80

# 查看文件信息
.\webptools.exe info myfile.webp

# 打开查看器
.\webptools.exe view animation.webp

# 比较两个动画
.\webptools.exe diff anim1.webp anim2.webp

# 查看完整帮助
.\webptools.exe help
```

## 📋 功能对比

| 功能 | 标准版 | 嵌入版 |
|-----|-------|-------|
| WebP压缩 | ✅ | ✅ |
| GIF转WebP | ❌ | ✅ |
| 文件信息查看 | ❌ | ✅ |
| WebP查看器 | ❌ | ✅ |
| 动画比较 | ❌ | ✅ |
| 质量分析 | ❌ | ✅ |
| 失真测量 | ❌ | ✅ |
| 信息转储 | ❌ | ✅ |
| 外部依赖 | 需要 | 无需 |
| 文件大小 | ~200KB | ~5MB |

## 🛠️ 内置工具（嵌入版）

嵌入版包含以下12个工具：

1. **webpmux.exe** - WebP动画信息解析和处理
2. **cwebp.exe** - 图像转WebP压缩工具
3. **dwebp.exe** - WebP解码工具
4. **gif2webp.exe** - GIF转WebP动画工具
5. **webpinfo.exe** - WebP文件信息查看器
6. **anim_diff.exe** - 动画比较工具
7. **anim_dump.exe** - 动画信息转储工具
8. **get_disto.exe** - 图像失真测量工具
9. **img2webp.exe** - 多图片转WebP动画
10. **webp_quality.exe** - WebP质量分析工具
11. **vwebp.exe** - WebP查看器
12. **freeglut.dll** - OpenGL库(vwebp依赖)

## 💡 使用建议

### 压缩质量建议
- **质量 10-30**: 极高压缩率，适合要求小文件的场景
- **质量 30-50**: 平衡压缩，推荐使用 ⭐
- **质量 50-80**: 高质量，文件稍大
- **质量 80-100**: 近乎无损，文件最大

### 选择哪个版本？

#### 选择标准版，如果你：
- 已经安装了libwebp工具
- 只需要基本的压缩功能
- 希望程序文件小巧

#### 选择嵌入版，如果你：
- 想要完整的WebP工具集 ⭐
- 不想安装外部依赖
- 需要多种WebP处理功能
- 要分发给其他人使用

## 🏗️ 项目结构

```
webpcompressor/
├── main.go                 # 标准版主程序
├── cmd/embedded/main.go    # 嵌入版主程序
├── embedded/               # 嵌入的二进制工具
│   ├── webpmux.exe
│   ├── cwebp.exe
│   └── ...其他工具
├── build.bat              # 构建脚本
├── go.mod                 # Go模块文件
└── README.md              # 说明文档
```

## 🔍 技术实现

### 标准版特性
- 使用系统已安装的libwebp工具
- 实现了完整的动画帧处理逻辑
- 支持高级压缩参数调整
- 修复了Go语言slice更新的经典陷阱

### 嵌入版特性
- 使用Go 1.16+ 的 `embed` 功能
- 运行时提取工具到临时目录
- 自动清理临时文件
- 支持中英文命令

## 🚨 注意事项

1. **嵌入版首次运行**会提取工具到临时目录（约1-2秒）
2. **标准版需要PATH中有webpmux和cwebp**
3. 建议使用**PowerShell或CMD**运行
4. 大文件处理时请确保有足够的**磁盘空间**

## 📈 性能对比

实测压缩效果（5.4MB动画文件）：

| 质量设置 | 输出大小 | 压缩率 | 质量评价 |
|---------|----------|--------|----------|
| 10      | 245KB    | 4.5%   | 高压缩，质量较低 |
| 30      | 863KB    | 15.5%  | 推荐设置 ⭐ |
| 50      | 1.2MB    | 22.1%  | 质量较好 |
| 80      | 2.8MB    | 51.8%  | 高质量 |

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个工具！

## 📄 许可证

MIT License - 可自由使用和修改。

---

**快速体验：** `.\webptools.exe help` 查看所有功能！ 