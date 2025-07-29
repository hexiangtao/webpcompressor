@echo off
echo 🎯 WebP压缩基础示例
echo ========================

echo.
echo 🔧 检查工具是否存在...
if not exist "..\..\bin\webpcompressor.exe" (
    echo ❌ 标准版工具不存在，请先运行构建脚本
    echo 💡 运行: .\build.bat
    goto end
)

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
    echo 💡 或者使用你自己的WebP文件替换下面的路径
    echo.
    echo 使用方法示例:
    echo   标准版: ..\..\bin\webpcompressor.exe your_file.webp 30 output.webp
    echo   嵌入版: ..\..\bin\webptools.exe compress your_file.webp 30 output.webp
    goto end
)

echo.
echo 🎬 开始压缩示例...

echo.
echo ==================== 标准版示例 ====================
echo 📝 使用标准版压缩（质量30）
for %%f in (..\sample_files\*.webp) do (
    echo 🔄 压缩文件: %%f
    ..\..\bin\webpcompressor.exe "%%f" 30 "..\output\%%~nf_standard_30.webp"
    if %ERRORLEVEL% EQU 0 (
        echo ✅ 完成: %%~nf_standard_30.webp
    ) else (
        echo ❌ 失败: %%f
    )
    echo.
)

echo.
echo ==================== 嵌入版示例 ====================
echo 📝 使用嵌入版压缩（质量30）
for %%f in (..\sample_files\*.webp) do (
    echo 🔄 压缩文件: %%f
    ..\..\bin\webptools.exe compress "%%f" 30 "..\output\%%~nf_embedded_30.webp"
    if %ERRORLEVEL% EQU 0 (
        echo ✅ 完成: %%~nf_embedded_30.webp
    ) else (
        echo ❌ 失败: %%f
    )
    echo.
)

echo.
echo ==================== 质量对比 ====================
echo 📝 不同质量设置对比（使用第一个文件）
for %%f in (..\sample_files\*.webp) do (
    echo 📊 质量对比示例（文件: %%f）
    
    echo   🔹 质量20（高压缩）
    ..\..\bin\webptools.exe compress "%%f" 20 "..\output\%%~nf_quality_20.webp"
    
    echo   🔹 质量50（平衡）
    ..\..\bin\webptools.exe compress "%%f" 50 "..\output\%%~nf_quality_50.webp"
    
    echo   🔹 质量80（高质量）
    ..\..\bin\webptools.exe compress "%%f" 80 "..\output\%%~nf_quality_80.webp"
    
    echo.
    echo 💡 查看输出文件大小对比:
    if exist "..\output\%%~nf_quality_20.webp" dir "..\output\%%~nf_quality_*.webp" | findstr "%%~nf_quality"
    
    goto quality_done
)
:quality_done

:end
echo.
echo 📂 输出文件位置: ..\output\
echo 💡 提示: 可以使用 ..\..\bin\webptools.exe info 命令查看文件详细信息
echo.
pause 