# 测试数据目录 (testdata/)

此目录存放项目的测试数据文件，遵循Go语言的标准约定。

## 目录结构

```
testdata/
├── input/          # 测试输入文件
│   ├── small.webp  # 小型测试文件
│   ├── large.webp  # 大型测试文件
│   └── animated.webp # 动画测试文件
├── expected/       # 预期输出结果
│   └── ...
├── output/         # 测试生成的输出文件（会被gitignore）
└── config/         # 测试配置文件
    └── test.yaml
```

## 使用说明

1. **输入文件** - 放在 `input/` 子目录中
2. **预期结果** - 放在 `expected/` 子目录中  
3. **临时输出** - 自动生成到 `output/` 子目录中（不会提交到Git）

## 测试约定

- 文件名应该描述性强，如：`small_animation.webp`
- 包含各种场景：不同大小、不同质量、不同帧数
- 输出文件会自动清理，无需手动管理

## 示例用法

```bash
# 使用测试数据进行压缩测试
bin\webpcompressor.exe testdata\input\animated.webp 30 testdata\output\result.webp

# 批量测试
bin\webptools.exe compress testdata\input\*.webp 40 testdata\output\
```

## 注意

- `output/` 目录中的文件是临时的，不会提交到版本控制
- 添加新测试文件时，请确保文件大小合理
- 测试文件应该覆盖各种使用场景 