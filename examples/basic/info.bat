@echo off
echo 📊 WebP文件信息查看示例
echo ===========================

echo.
echo 🔧 检查工具是否存在...
if not exist "..\..\bin\webptools.exe" (
    echo ❌ 嵌入版工具不存在，请先运行构建脚本
    echo 💡 运行: .\build.bat
    goto end
)

echo ✅ 工具检查完成

echo.
echo 📁 检查示例文件...
if not exist "..\sample_files\*.webp" (
    echo ⚠️  示例文件不存在，请将WebP文件放入 ..\sample_files\ 目录
    echo 💡 或者使用你自己的WebP文件
    echo.
    echo 使用方法示例:
    echo   ..\..\bin\webptools.exe info your_file.webp
    goto end
)

echo.
echo 🔍 开始分析WebP文件...

echo.
echo ==================== 文件信息分析 ====================
for %%f in (..\sample_files\*.webp) do (
    echo.
    echo 📄 分析文件: %%f
    echo ----------------------------------------
    ..\..\bin\webptools.exe info "%%f"
    echo ----------------------------------------
    echo.
)

echo.
echo ==================== 输出文件分析 ====================
if exist "..\output\*.webp" (
    echo 📊 分析压缩后的输出文件...
    for %%f in (..\output\*_quality_*.webp) do (
        echo.
        echo 📄 分析输出文件: %%f
        echo ----------------------------------------
        ..\..\bin\webptools.exe info "%%f"
        echo ----------------------------------------
        echo.
    )
) else (
    echo ⚠️  没有找到输出文件，请先运行压缩示例
    echo 💡 运行: compress.bat
)

echo.
echo ==================== 文件大小对比 ====================
echo 📈 文件大小统计:
echo.

if exist "..\sample_files\*.webp" (
    echo 🔵 原始文件:
    for %%f in (..\sample_files\*.webp) do (
        for /f "tokens=3" %%a in ('dir "%%f" 2^>nul ^| findstr /r "[0-9]"') do (
            echo   %%~nxf: %%a bytes
        )
    )
)

echo.
if exist "..\output\*_quality_*.webp" (
    echo 🟢 压缩文件:
    for %%f in (..\output\*_quality_*.webp) do (
        for /f "tokens=3" %%a in ('dir "%%f" 2^>nul ^| findstr /r "[0-9]"') do (
            echo   %%~nxf: %%a bytes
        )
    )
)

:end
echo.
echo 📋 总结:
echo   • 使用 info 命令可以查看WebP文件的详细信息
echo   • 包括画布大小、帧数、循环次数等
echo   • 可以用来验证压缩效果和文件完整性
echo.
pause 