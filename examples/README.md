# 示例目录 (examples/)

此目录包含使用示例和演示文件。

## 目录结构

```
examples/
├── basic/              # 基础使用示例
│   ├── compress.bat    # 压缩示例脚本
│   └── info.bat        # 信息查看示例
├── advanced/           # 高级使用示例
│   ├── batch_process.bat # 批量处理示例
│   └── quality_test.bat  # 质量对比示例
├── sample_files/       # 示例文件
│   ├── small_anim.webp # 小动画示例
│   └── large_anim.webp # 大动画示例
└── output/             # 示例输出目录（会被gitignore）
```

## 使用说明

### 基础示例

1. **压缩WebP动画**
   ```bash
   cd examples\basic
   compress.bat
   ```

2. **查看WebP信息**
   ```bash
   cd examples\basic
   info.bat
   ```

### 高级示例

1. **批量处理**
   ```bash
   cd examples\advanced
   batch_process.bat
   ```

2. **质量对比测试**
   ```bash
   cd examples\advanced
   quality_test.bat
   ```

## 示例脚本说明

- **compress.bat** - 演示基本压缩功能
- **info.bat** - 演示信息查看功能
- **batch_process.bat** - 演示批量处理多个文件
- **quality_test.bat** - 演示不同质量设置的效果对比

## 注意

- 示例脚本假设已经构建了工具到 `bin/` 目录
- `output/` 目录用于存放示例运行结果，不会提交到Git
- 示例文件应保持较小的大小以便快速演示

## 自定义示例

你可以：
1. 将自己的WebP文件放到 `sample_files/` 目录
2. 修改示例脚本以使用你的文件
3. 运行示例查看效果 