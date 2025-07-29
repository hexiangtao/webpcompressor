@echo off
echo ⚡ WebP批量处理高级示例
echo ============================

echo.
echo 🔧 检查工具是否存在...
if not exist "..\..\bin\webptools.exe" (
    echo ❌ 嵌入版工具不存在，请先运行构建脚本
    echo 💡 运行: .\build.bat
    goto end
)

echo ✅ 工具检查完成

echo.
echo 📋 批量处理选项:
echo   1. 批量压缩（多种质量）
echo   2. 批量分析文件信息
echo   3. 性能基准测试
echo   4. 全部执行
echo.

set /p choice="请选择处理方式 (1/2/3/4): "

if "%choice%"=="1" goto batch_compress
if "%choice%"=="2" goto batch_info
if "%choice%"=="3" goto benchmark
if "%choice%"=="4" goto process_all
echo ❌ 无效选择
goto end

:batch_compress
echo.
echo ==================== 批量压缩 ====================
echo 📦 为每个文件生成多种质量版本...

if not exist "..\sample_files\*.webp" (
    echo ⚠️  示例文件不存在，请将WebP文件放入 ..\sample_files\ 目录
    goto end
)

set "qualities=20 30 40 50 60 70 80"

for %%f in (..\sample_files\*.webp) do (
    echo.
    echo 🎞️  处理文件: %%~nxf
    echo ─────────────────────────────────────
    
    for %%q in (%qualities%) do (
        echo 🔄 生成质量%%q版本...
        ..\..\bin\webptools.exe compress "%%f" %%q "..\output\%%~nf_batch_%%q.webp"
        if %ERRORLEVEL% EQU 0 (
            echo   ✅ 完成: %%~nf_batch_%%q.webp
        ) else (
            echo   ❌ 失败: 质量%%q
        )
    )
)

echo.
echo 📊 批量压缩统计:
if exist "..\output\*_batch_*.webp" (
    for /f %%i in ('dir /b "..\output\*_batch_*.webp" 2^>nul ^| find /c /v ""') do (
        echo   ✅ 成功生成 %%i 个压缩文件
    )
)
goto process_done

:batch_info
echo.
echo ==================== 批量信息分析 ====================
echo 📊 分析所有WebP文件的详细信息...

echo.
echo 🔵 原始文件分析:
for %%f in (..\sample_files\*.webp) do (
    echo.
    echo 📄 %%~nxf
    echo ─────────────────────────────────────
    ..\..\bin\webptools.exe info "%%f"
    echo.
)

echo.
echo 🟢 输出文件分析:
if exist "..\output\*.webp" (
    for %%f in (..\output\*.webp) do (
        echo.
        echo 📄 %%~nxf
        echo ─────────────────────────────────────
        ..\..\bin\webptools.exe info "%%f"
        echo.
    )
) else (
    echo ⚠️  没有找到输出文件
)
goto process_done

:benchmark
echo.
echo ==================== 性能基准测试 ====================
echo ⏱️  测试不同质量设置的处理时间...

if not exist "..\sample_files\*.webp" (
    echo ⚠️  示例文件不存在，请将WebP文件放入 ..\sample_files\ 目录
    goto end
)

set "test_qualities=10 30 50 80"

for %%f in (..\sample_files\*.webp) do (
    echo.
    echo 🧪 基准测试文件: %%~nxf
    echo ═══════════════════════════════════════
    
    for %%q in (%test_qualities%) do (
        echo.
        echo 🔬 测试质量 %%q...
        echo 开始时间: %time%
        
        ..\..\bin\webptools.exe compress "%%f" %%q "..\output\%%~nf_bench_%%q.webp"
        
        echo 结束时间: %time%
        
        if exist "..\output\%%~nf_bench_%%q.webp" (
            for /f "tokens=3" %%s in ('dir "..\output\%%~nf_bench_%%q.webp" 2^>nul ^| findstr /r "[0-9]"') do (
                echo 输出大小: %%s bytes
            )
        )
        echo ─────────────────────────────────────
    )
    
    goto benchmark_done
)
:benchmark_done
goto process_done

:process_all
echo.
echo ==================== 执行全部流程 ====================
call :batch_compress
call :batch_info
call :benchmark
goto process_done

:process_done
echo.
echo ==================== 处理完成 ====================
echo 📂 输出目录: ..\output\
echo 📋 文件统计:

if exist "..\output\*.webp" (
    for /f %%i in ('dir /b "..\output\*.webp" 2^>nul ^| find /c /v ""') do (
        echo   📁 总计生成文件: %%i 个
    )
    
    echo.
    echo 📊 文件大小分布:
    for %%f in (..\output\*.webp) do (
        for /f "tokens=3" %%s in ('dir "%%f" 2^>nul ^| findstr /r "[0-9]"') do (
            echo   📄 %%~nxf: %%s bytes
        )
    )
) else (
    echo   ⚠️  没有生成输出文件
)

echo.
echo 💡 后续操作建议:
echo   • 使用 info.bat 查看详细文件信息
echo   • 比较不同质量设置的效果
echo   • 根据需求选择最佳质量参数

:end
echo.
pause 