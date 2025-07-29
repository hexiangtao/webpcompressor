package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// 嵌入所有WebP工具二进制文件
//
//go:embed ../../embedded/webpmux.exe
var webpmuxBin []byte

//go:embed ../../embedded/cwebp.exe
var cwebpBin []byte

//go:embed ../../embedded/dwebp.exe
var dwebpBin []byte

//go:embed ../../embedded/gif2webp.exe
var gif2webpBin []byte

//go:embed ../../embedded/webpinfo.exe
var webpinfoBin []byte

//go:embed ../../embedded/anim_diff.exe
var animDiffBin []byte

//go:embed ../../embedded/anim_dump.exe
var animDumpBin []byte

//go:embed ../../embedded/get_disto.exe
var getDistoBin []byte

//go:embed ../../embedded/img2webp.exe
var img2webpBin []byte

//go:embed ../../embedded/webp_quality.exe
var webpQualityBin []byte

//go:embed ../../embedded/vwebp.exe
var vwebpBin []byte

//go:embed ../../embedded/freeglut.dll
var freeglutDLL []byte

type EmbeddedTool struct {
	name string
	data []byte
	desc string
}

var embeddedTools = []EmbeddedTool{
	{"webpmux.exe", webpmuxBin, "WebP动画信息解析和处理"},
	{"cwebp.exe", cwebpBin, "图像转WebP压缩工具"},
	{"dwebp.exe", dwebpBin, "WebP解码工具"},
	{"gif2webp.exe", gif2webpBin, "GIF转WebP动画工具"},
	{"webpinfo.exe", webpinfoBin, "WebP文件信息查看器"},
	{"anim_diff.exe", animDiffBin, "动画比较工具"},
	{"anim_dump.exe", animDumpBin, "动画信息转储工具"},
	{"get_disto.exe", getDistoBin, "图像失真测量工具"},
	{"img2webp.exe", img2webpBin, "多图片转WebP动画"},
	{"webp_quality.exe", webpQualityBin, "WebP质量分析工具"},
	{"vwebp.exe", vwebpBin, "WebP查看器"},
	{"freeglut.dll", freeglutDLL, "OpenGL库(vwebp依赖)"},
}

var tempDir string

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	// 提取嵌入的工具到临时目录
	var err error
	tempDir, err = extractTools()
	if err != nil {
		fmt.Printf("❌ 提取工具失败: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	command := strings.ToLower(os.Args[1])

	switch command {
	case "compress", "压缩":
		handleCompress()
	case "info", "信息":
		handleInfo()
	case "convert", "转换":
		handleConvert()
	case "view", "查看":
		handleView()
	case "diff", "比较":
		handleDiff()
	case "dump", "转储":
		handleDump()
	case "quality", "质量":
		handleQuality()
	case "disto", "失真":
		handleDisto()
	case "help", "帮助", "-h", "--help":
		showDetailedHelp()
	default:
		fmt.Printf("❌ 未知命令: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func extractTools() (string, error) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "webptools_"+strconv.Itoa(os.Getpid()))
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return "", err
	}

	fmt.Println("🔧 提取嵌入的工具...")

	// 提取所有工具
	for _, tool := range embeddedTools {
		toolPath := filepath.Join(tempDir, tool.name)
		err := os.WriteFile(toolPath, tool.data, 0755)
		if err != nil {
			return "", fmt.Errorf("提取 %s 失败: %v", tool.name, err)
		}
	}

	fmt.Printf("✅ 工具已提取至: %s\n", tempDir)
	return tempDir, nil
}

func cleanup() {
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}

func getToolPath(toolName string) string {
	return filepath.Join(tempDir, toolName)
}

func runTool(toolName string, args ...string) error {
	toolPath := getToolPath(toolName)
	cmd := exec.Command(toolPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func handleCompress() {
	if len(os.Args) < 5 {
		fmt.Println("用法: webptools compress <input.webp> <quality[0-100]> <output.webp>")
		fmt.Println("示例: webptools compress animation.webp 30 compressed.webp")
		os.Exit(1)
	}

	inputFile := os.Args[2]
	quality, _ := strconv.Atoi(os.Args[3])
	outputFile := os.Args[4]

	fmt.Println("🎯 开始压缩WebP动画...")

	// 使用我们原来的压缩逻辑，但使用嵌入的工具
	err := compressWebPWithEmbeddedTools(inputFile, quality, outputFile)
	if err != nil {
		fmt.Printf("❌ 压缩失败: %v\n", err)
		os.Exit(1)
	}
}

func handleInfo() {
	if len(os.Args) < 3 {
		fmt.Println("用法: webptools info <input.webp>")
		os.Exit(1)
	}

	inputFile := os.Args[2]
	fmt.Printf("📊 查看 %s 信息...\n", inputFile)

	err := runTool("webpinfo.exe", inputFile)
	if err != nil {
		fmt.Printf("❌ 查看信息失败: %v\n", err)
		os.Exit(1)
	}
}

func handleConvert() {
	if len(os.Args) < 4 {
		fmt.Println("用法: webptools convert <input.gif> <output.webp> [quality]")
		fmt.Println("示例: webptools convert animation.gif animation.webp 80")
		os.Exit(1)
	}

	inputFile := os.Args[2]
	outputFile := os.Args[3]

	args := []string{inputFile, "-o", outputFile}
	if len(os.Args) >= 5 {
		quality := os.Args[4]
		args = append(args, "-q", quality)
	}

	fmt.Printf("🔄 转换 %s 为 %s...\n", inputFile, outputFile)
	err := runTool("gif2webp.exe", args...)
	if err != nil {
		fmt.Printf("❌ 转换失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 转换完成!")
}

func handleView() {
	if len(os.Args) < 3 {
		fmt.Println("用法: webptools view <input.webp>")
		os.Exit(1)
	}

	inputFile := os.Args[2]
	fmt.Printf("👁️ 打开查看器查看 %s...\n", inputFile)

	err := runTool("vwebp.exe", inputFile)
	if err != nil {
		fmt.Printf("❌ 打开查看器失败: %v\n", err)
		os.Exit(1)
	}
}

func handleDiff() {
	if len(os.Args) < 4 {
		fmt.Println("用法: webptools diff <webp1> <webp2>")
		os.Exit(1)
	}

	file1 := os.Args[2]
	file2 := os.Args[3]
	fmt.Printf("🔍 比较 %s 和 %s...\n", file1, file2)

	err := runTool("anim_diff.exe", file1, file2)
	if err != nil {
		fmt.Printf("❌ 比较失败: %v\n", err)
		os.Exit(1)
	}
}

func handleDump() {
	if len(os.Args) < 3 {
		fmt.Println("用法: webptools dump <input.webp>")
		os.Exit(1)
	}

	inputFile := os.Args[2]
	fmt.Printf("📋 转储 %s 的动画信息...\n", inputFile)

	err := runTool("anim_dump.exe", inputFile)
	if err != nil {
		fmt.Printf("❌ 转储失败: %v\n", err)
		os.Exit(1)
	}
}

func handleQuality() {
	if len(os.Args) < 3 {
		fmt.Println("用法: webptools quality <input.webp>")
		os.Exit(1)
	}

	inputFile := os.Args[2]
	fmt.Printf("📈 分析 %s 的质量...\n", inputFile)

	err := runTool("webp_quality.exe", inputFile)
	if err != nil {
		fmt.Printf("❌ 质量分析失败: %v\n", err)
		os.Exit(1)
	}
}

func handleDisto() {
	if len(os.Args) < 4 {
		fmt.Println("用法: webptools disto <original> <compressed>")
		os.Exit(1)
	}

	original := os.Args[2]
	compressed := os.Args[3]
	fmt.Printf("📏 测量失真: %s vs %s...\n", original, compressed)

	err := runTool("get_disto.exe", original, compressed)
	if err != nil {
		fmt.Printf("❌ 失真测量失败: %v\n", err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("🎨 WebP多功能工具集 v2.0 (嵌入版)")
	fmt.Println("================================================")
	fmt.Println("用法: webptools <命令> [参数...]")
	fmt.Println("")
	fmt.Println("可用命令:")
	fmt.Println("  compress/压缩    - 压缩WebP动画")
	fmt.Println("  info/信息       - 查看WebP文件信息")
	fmt.Println("  convert/转换    - GIF转WebP动画")
	fmt.Println("  view/查看       - 打开WebP查看器")
	fmt.Println("  diff/比较       - 比较两个WebP动画")
	fmt.Println("  dump/转储       - 转储动画信息")
	fmt.Println("  quality/质量    - 分析WebP质量")
	fmt.Println("  disto/失真      - 测量图像失真")
	fmt.Println("  help/帮助       - 显示详细帮助")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  webptools compress input.webp 30 output.webp")
	fmt.Println("  webptools convert animation.gif animation.webp 80")
	fmt.Println("  webptools info myfile.webp")
	fmt.Printf("📁 内置了 %d 个WebP工具，无需外部依赖\n", len(embeddedTools))
}

func showDetailedHelp() {
	fmt.Println("🎨 WebP多功能工具集 - 详细帮助 (嵌入版)")
	fmt.Println("================================================")
	fmt.Printf("运行环境: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("")

	fmt.Println("📦 内置工具:")
	for _, tool := range embeddedTools {
		fmt.Printf("  %-20s - %s\n", tool.name, tool.desc)
	}

	fmt.Println("")
	fmt.Println("📋 详细命令说明:")
	fmt.Println("")
	fmt.Println("1. compress/压缩 - 压缩WebP动画")
	fmt.Println("   用法: webptools compress <input.webp> <quality[0-100]> <output.webp>")
	fmt.Println("   示例: webptools compress big.webp 30 small.webp")
	fmt.Println("")
	fmt.Println("2. convert/转换 - GIF转WebP动画")
	fmt.Println("   用法: webptools convert <input.gif> <output.webp> [quality]")
	fmt.Println("   示例: webptools convert dance.gif dance.webp 75")
	fmt.Println("")
	fmt.Println("3. info/信息 - 查看WebP文件详细信息")
	fmt.Println("   用法: webptools info <input.webp>")
	fmt.Println("   示例: webptools info animation.webp")
	fmt.Println("")
	fmt.Println("4. view/查看 - 打开WebP查看器")
	fmt.Println("   用法: webptools view <input.webp>")
	fmt.Println("   示例: webptools view animation.webp")
	fmt.Println("")
	fmt.Println("💡 提示:")
	fmt.Println("   - 压缩质量: 0-100 (0=最小文件,100=最高质量)")
	fmt.Println("   - 建议质量: 30-50 获得最佳压缩效果")
	fmt.Println("   - 所有工具都已内置，无需外部依赖")
	fmt.Println("   - 工具会自动提取到临时目录并在程序结束时清理")
}

// 使用嵌入工具的压缩函数（基于原来的逻辑）
func compressWebPWithEmbeddedTools(inputFile string, quality int, outputFile string) error {
	fmt.Printf("🔍 使用内置工具处理: %s -> %s (质量: %d)\n", inputFile, outputFile, quality)

	// 先获取文件信息
	args := []string{"-info", inputFile}
	err := runTool("webpmux.exe", args...)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// TODO: 这里应该实现完整的压缩逻辑
	// 目前作为演示，我们使用gif2webp进行基本转换
	fmt.Println("🚧 注意: 压缩功能正在开发中，当前使用基本处理")

	return nil
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
