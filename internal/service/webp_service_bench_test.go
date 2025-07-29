package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"webpcompressor/internal/domain"
)

// BenchmarkParseAnimationLarge 基准测试解析大型动画
func BenchmarkParseAnimationLarge(b *testing.B) {
	service := createTestWebPService()
	mockToolExecutor := service.toolExecutor.(*MockToolExecutor)

	// 模拟包含100帧的大型动画
	mockOutput := generateMockAnimationOutput(100)
	mockToolExecutor.SetMockOutput("webpmux -info test.webp", mockOutput)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ParseAnimation(ctx, "test.webp")
		if err != nil {
			b.Fatalf("ParseAnimation failed: %v", err)
		}
	}
}

// BenchmarkCompressFramesSequential 基准测试顺序压缩
func BenchmarkCompressFramesSequential(b *testing.B) {
	benchmarkCompressFrames(b, false, "sequential")
}

// BenchmarkCompressFramesParallel 基准测试并行压缩
func BenchmarkCompressFramesParallel(b *testing.B) {
	benchmarkCompressFrames(b, true, "parallel")
}

// benchmarkCompressFrames 压缩帧基准测试通用函数
func benchmarkCompressFrames(b *testing.B, enableParallel bool, name string) {
	service := createTestWebPService()
	mockFileManager := service.fileManager.(*MockFileManager)

	// 创建测试帧
	frameCount := 50
	frames := make([]*domain.FrameInfo, frameCount)
	for i := 0; i < frameCount; i++ {
		frames[i] = &domain.FrameInfo{
			Index: i + 1,
			Path:  fmt.Sprintf("frame_%d.webp", i+1),
		}
		mockFileManager.SetFileExists(frames[i].Path, true)
	}

	config := domain.DefaultCompressionConfig(50)
	config.EnableParallel = enableParallel
	config.MaxConcurrency = 4

	ctx := context.Background()

	b.Run(name, func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 重置帧路径
			for j := 0; j < frameCount; j++ {
				frames[j].Path = fmt.Sprintf("frame_%d.webp", j+1)
			}

			err := service.CompressFrames(ctx, frames, config)
			if err != nil {
				b.Fatalf("CompressFrames failed: %v", err)
			}
		}
	})
}

// BenchmarkDifferentFrameCounts 测试不同帧数的性能
func BenchmarkDifferentFrameCounts(b *testing.B) {
	frameCounts := []int{10, 50, 100, 200}

	for _, frameCount := range frameCounts {
		b.Run(fmt.Sprintf("frames_%d_sequential", frameCount), func(b *testing.B) {
			benchmarkWithFrameCount(b, frameCount, false)
		})

		b.Run(fmt.Sprintf("frames_%d_parallel", frameCount), func(b *testing.B) {
			benchmarkWithFrameCount(b, frameCount, true)
		})
	}
}

// benchmarkWithFrameCount 指定帧数的基准测试
func benchmarkWithFrameCount(b *testing.B, frameCount int, enableParallel bool) {
	service := createTestWebPService()
	mockFileManager := service.fileManager.(*MockFileManager)

	// 创建测试帧
	frames := make([]*domain.FrameInfo, frameCount)
	for i := 0; i < frameCount; i++ {
		frames[i] = &domain.FrameInfo{
			Index: i + 1,
			Path:  fmt.Sprintf("frame_%d.webp", i+1),
		}
		mockFileManager.SetFileExists(frames[i].Path, true)
	}

	config := domain.DefaultCompressionConfig(50)
	config.EnableParallel = enableParallel
	config.MaxConcurrency = 4

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 重置帧路径
		for j := 0; j < frameCount; j++ {
			frames[j].Path = fmt.Sprintf("frame_%d.webp", j+1)
		}

		err := service.CompressFrames(ctx, frames, config)
		if err != nil {
			b.Fatalf("CompressFrames failed: %v", err)
		}
	}
}

// BenchmarkDifferentQualitySettings 测试不同质量设置的性能
func BenchmarkDifferentQualitySettings(b *testing.B) {
	qualities := []int{10, 30, 50, 75, 90}

	for _, quality := range qualities {
		b.Run(fmt.Sprintf("quality_%d", quality), func(b *testing.B) {
			benchmarkWithQuality(b, quality)
		})
	}
}

// benchmarkWithQuality 指定质量的基准测试
func benchmarkWithQuality(b *testing.B, quality int) {
	service := createTestWebPService()
	mockFileManager := service.fileManager.(*MockFileManager)

	// 创建测试帧
	frameCount := 20
	frames := make([]*domain.FrameInfo, frameCount)
	for i := 0; i < frameCount; i++ {
		frames[i] = &domain.FrameInfo{
			Index: i + 1,
			Path:  fmt.Sprintf("frame_%d.webp", i+1),
		}
		mockFileManager.SetFileExists(frames[i].Path, true)
	}

	config := domain.DefaultCompressionConfig(quality)
	config.EnableParallel = true
	config.MaxConcurrency = 4

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 重置帧路径
		for j := 0; j < frameCount; j++ {
			frames[j].Path = fmt.Sprintf("frame_%d.webp", j+1)
		}

		err := service.CompressFrames(ctx, frames, config)
		if err != nil {
			b.Fatalf("CompressFrames failed: %v", err)
		}
	}
}

// BenchmarkDifferentConcurrencyLevels 测试不同并发级别的性能
func BenchmarkDifferentConcurrencyLevels(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			benchmarkWithConcurrency(b, concurrency)
		})
	}
}

// benchmarkWithConcurrency 指定并发级别的基准测试
func benchmarkWithConcurrency(b *testing.B, concurrency int) {
	service := createTestWebPService()
	mockFileManager := service.fileManager.(*MockFileManager)

	// 创建测试帧
	frameCount := 50
	frames := make([]*domain.FrameInfo, frameCount)
	for i := 0; i < frameCount; i++ {
		frames[i] = &domain.FrameInfo{
			Index: i + 1,
			Path:  fmt.Sprintf("frame_%d.webp", i+1),
		}
		mockFileManager.SetFileExists(frames[i].Path, true)
	}

	config := domain.DefaultCompressionConfig(50)
	config.EnableParallel = concurrency > 1
	config.MaxConcurrency = concurrency

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 重置帧路径
		for j := 0; j < frameCount; j++ {
			frames[j].Path = fmt.Sprintf("frame_%d.webp", j+1)
		}

		err := service.CompressFrames(ctx, frames, config)
		if err != nil {
			b.Fatalf("CompressFrames failed: %v", err)
		}
	}
}

// BenchmarkMemoryUsage 内存使用基准测试
func BenchmarkMemoryUsage(b *testing.B) {
	service := createTestWebPService()
	mockToolExecutor := service.toolExecutor.(*MockToolExecutor)
	mockFileManager := service.fileManager.(*MockFileManager)

	// 模拟大型动画
	frameCount := 1000
	mockOutput := generateMockAnimationOutput(frameCount)
	mockToolExecutor.SetMockOutput("webpmux -info large.webp", mockOutput)

	// 设置文件存在
	for i := 1; i <= frameCount; i++ {
		mockFileManager.SetFileExists(fmt.Sprintf("frame_%d.webp", i), true)
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs() // 报告内存分配

	for i := 0; i < b.N; i++ {
		animInfo, err := service.ParseAnimation(ctx, "large.webp")
		if err != nil {
			b.Fatalf("ParseAnimation failed: %v", err)
		}

		config := domain.DefaultCompressionConfig(50)
		config.EnableParallel = true
		config.MaxConcurrency = 4

		err = service.CompressFrames(ctx, animInfo.Frames, config)
		if err != nil {
			b.Fatalf("CompressFrames failed: %v", err)
		}
	}
}

// BenchmarkFullCompressionWorkflow 完整压缩流程基准测试
func BenchmarkFullCompressionWorkflow(b *testing.B) {
	service := createTestWebPService()
	mockToolExecutor := service.toolExecutor.(*MockToolExecutor)
	mockFileManager := service.fileManager.(*MockFileManager)

	// 模拟中等大小的动画
	frameCount := 50
	mockOutput := generateMockAnimationOutput(frameCount)
	mockToolExecutor.SetMockOutput("webpmux -info input.webp", mockOutput)

	// 设置输入文件存在
	mockFileManager.SetFileExists("input.webp", true)
	mockFileManager.SetFileSize("input.webp", 5*1024*1024) // 5MB

	// 设置所有帧文件存在
	for i := 1; i <= frameCount; i++ {
		mockFileManager.SetFileExists(fmt.Sprintf("frame_%d.webp", i), true)
		mockFileManager.SetFileExists(fmt.Sprintf("frame_compressed_%d.webp", i), true)
	}

	// 设置输出文件
	mockFileManager.SetFileExists("output.webp", true)
	mockFileManager.SetFileSize("output.webp", 2*1024*1024) // 2MB

	config := domain.DefaultCompressionConfig(50)
	config.EnableParallel = true
	config.MaxConcurrency = 4

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CompressAnimation(ctx, "input.webp", "output.webp", config)
		if err != nil {
			b.Fatalf("CompressAnimation failed: %v", err)
		}
	}
}

// generateMockAnimationOutput 生成模拟的动画信息输出
func generateMockAnimationOutput(frameCount int) string {
	var builder strings.Builder

	builder.WriteString("Canvas size: 1920 x 1080\n")
	builder.WriteString("Features present: animation transparency\n")
	builder.WriteString("Background color : 0xFFFFFFFF\n")
	builder.WriteString("Loop Count : 0\n")
	builder.WriteString(fmt.Sprintf("Number of frames: %d\n", frameCount))
	builder.WriteString("No.: width height alpha x_offset y_offset duration dispose blend image_size compression\n")

	for i := 1; i <= frameCount; i++ {
		builder.WriteString(fmt.Sprintf("%3d:   1920   1080   yes         0        0       33    none    no     %d      lossy\n",
			i, 50000+i*100))
	}

	return builder.String()
}

// 性能比较测试
func BenchmarkPerformanceComparison(b *testing.B) {
	testCases := []struct {
		name        string
		frameCount  int
		parallel    bool
		concurrency int
		quality     int
	}{
		{"small_sequential", 10, false, 1, 50},
		{"small_parallel", 10, true, 4, 50},
		{"medium_sequential", 50, false, 1, 50},
		{"medium_parallel", 50, true, 4, 50},
		{"large_sequential", 100, false, 1, 50},
		{"large_parallel", 100, true, 4, 50},
		{"high_quality", 50, true, 4, 90},
		{"low_quality", 50, true, 4, 10},
		{"high_concurrency", 50, true, 8, 50},
		{"low_concurrency", 50, true, 2, 50},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			service := createTestWebPService()
			mockFileManager := service.fileManager.(*MockFileManager)

			// 创建测试帧
			frames := make([]*domain.FrameInfo, tc.frameCount)
			for i := 0; i < tc.frameCount; i++ {
				frames[i] = &domain.FrameInfo{
					Index: i + 1,
					Path:  fmt.Sprintf("frame_%d.webp", i+1),
				}
				mockFileManager.SetFileExists(frames[i].Path, true)
			}

			config := domain.DefaultCompressionConfig(tc.quality)
			config.EnableParallel = tc.parallel
			config.MaxConcurrency = tc.concurrency

			ctx := context.Background()

			b.ResetTimer()
			start := time.Now()

			for i := 0; i < b.N; i++ {
				// 重置帧路径
				for j := 0; j < tc.frameCount; j++ {
					frames[j].Path = fmt.Sprintf("frame_%d.webp", j+1)
				}

				err := service.CompressFrames(ctx, frames, config)
				if err != nil {
					b.Fatalf("CompressFrames failed: %v", err)
				}
			}

			duration := time.Since(start)
			b.ReportMetric(float64(tc.frameCount)/duration.Seconds()*float64(b.N), "frames/sec")
		})
	}
}
