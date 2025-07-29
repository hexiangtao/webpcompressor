package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type FrameInfo struct {
	Index     int
	X         int
	Y         int
	Duration  int
	Dispose   int
	Blend     int
	FramePath string
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("用法: webpcompressor <input.webp> <quality[0-100]> <output.webp>")
		fmt.Println("提示: 质量建议30-50可获得更好的压缩效果")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	quality, _ := strconv.Atoi(os.Args[2])
	outputFile := os.Args[3]

	tmpDir := "temp_frames"
	os.MkdirAll(tmpDir, 0755)

	fmt.Println("🔍 解析帧控制信息...")
	frames, err := parseFrameInfo(inputFile)
	if err != nil {
		fmt.Println("❌ 解析失败:", err)
		os.Exit(1)
	}

	fmt.Println("📦 提取每一帧...")
	for _, frame := range frames {
		frameOut := filepath.Join(tmpDir, fmt.Sprintf("frame_%d.webp", frame.Index))
		err := exec.Command("webpmux", "-get", "frame", strconv.Itoa(frame.Index), "-o", frameOut, inputFile).Run()
		if err != nil {
			fmt.Printf("❌ 提取第 %d 帧失败\n", frame.Index)
			os.Exit(1)
		}
		frames[frame.Index-1].FramePath = frameOut
	}

	fmt.Println("🎯 压缩帧图像...")
	for i := range frames {
		out := filepath.Join(tmpDir, fmt.Sprintf("frame_compressed_%d.webp", frames[i].Index))
		// 使用更激进的压缩设置
		args := []string{
			"-q", strconv.Itoa(quality), // 质量设置
			"-m", "6", // 压缩方法 (0-6, 6最慢但压缩率最高)
			"-preset", "photo", // 改为photo预设，更适合真实图像
			"-mt",       // 多线程
			"-f", "100", // 最大滤波强度
			"-sharpness", "0", // 锐化 (0-7, 0为最柔和)
			"-sns", "100", // 空间噪声整形 (0-100)
			"-segments", "4", // 分段数量 (1-4)
			"-pass", "10", // 分析遍数 (1-10)
			"-alpha_q", strconv.Itoa(quality / 2), // alpha通道质量更低
			"-size", "0", // 不限制文件大小，专注质量
			"-metadata", "none", // 移除元数据
			frames[i].FramePath,
			"-o", out,
		}
		err := exec.Command("cwebp", args...).Run()
		if err != nil {
			fmt.Printf("❌ 压缩第 %d 帧失败\n", frames[i].Index)
			os.Exit(1)
		}
		frames[i].FramePath = out // 现在这会正确更新原始slice
	}

	fmt.Println("🧩 重组帧动画...")

	args := []string{}
	for _, frame := range frames {
		// 正确的格式：+duration+x_offset+y_offset+dispose_method+blend_method
		blendStr := "-b" // no blending
		if frame.Blend == 1 {
			blendStr = "+b" // blending
		}
		frameArg := fmt.Sprintf("+%d+%d+%d+%d%s", frame.Duration, frame.X, frame.Y, frame.Dispose, blendStr)
		args = append(args, "-frame", frame.FramePath, frameArg)
	}
	args = append(args, "-loop", "0", "-o", outputFile)

	err = exec.Command("webpmux", args...).Run()
	if err != nil {
		fmt.Println("❌ 重组失败:", err)
		os.Exit(1)
	}

	// 显示压缩效果
	inputStat, _ := os.Stat(inputFile)
	outputStat, _ := os.Stat(outputFile)
	inputSize := inputStat.Size()
	outputSize := outputStat.Size()
	ratio := float64(outputSize) / float64(inputSize) * 100

	fmt.Printf("✅ 完成！压缩动图输出至: %s\n", outputFile)
	fmt.Printf("📊 压缩效果: %s -> %s (%.1f%%)\n",
		formatFileSize(inputSize), formatFileSize(outputSize), ratio)

	// 清理临时文件
	fmt.Println("🧹 清理临时文件...")
	os.RemoveAll(tmpDir)
}
func parseFrameInfo(input string) ([]FrameInfo, error) {
	cmd := exec.Command("webpmux", "-info", input)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	startReading := false
	var frames []FrameInfo

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 检测到表头开始
		if strings.HasPrefix(line, "No.") && strings.Contains(line, "duration") {
			startReading = true
			continue
		}

		if startReading {
			if line == "" {
				break
			}
			fields := strings.Fields(line)
			if len(fields) >= 9 {
				// 实际格式：No.: width height alpha x_offset y_offset duration dispose blend image_size compression
				indexStr := strings.TrimSuffix(fields[0], ":") // 移除冒号
				index, _ := strconv.Atoi(indexStr)
				x, _ := strconv.Atoi(fields[4])        // x_offset
				y, _ := strconv.Atoi(fields[5])        // y_offset
				duration, _ := strconv.Atoi(fields[6]) // duration

				// 处理dispose字段（"none"/"background" -> 0/1）
				dispose := 0
				if fields[7] == "background" {
					dispose = 1
				}

				// 处理blend字段（"no"/"yes" -> 0/1）
				blend := 0
				if fields[8] == "yes" {
					blend = 1
				}

				frames = append(frames, FrameInfo{
					Index:    index,
					X:        x,
					Y:        y,
					Duration: duration,
					Dispose:  dispose,
					Blend:    blend,
				})
			}
		}
	}

	if len(frames) == 0 {
		return nil, fmt.Errorf("未能解析到任何帧")
	}
	return frames, nil
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
