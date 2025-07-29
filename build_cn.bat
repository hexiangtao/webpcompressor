@echo off
chcp 936 > nul
echo WebP工具构建脚本
echo =======================================

echo.
echo 构建选项:
echo   1. 标准版 (需要外部工具) - bin\webpcompressor.exe
echo   2. 嵌入版 (内置所有工具) - bin\webptools.exe  
echo   3. 同时构建两个版本
echo.

set /p choice="请选择构建选项 (1/2/3): "

if "%choice%"=="1" goto build_standard
if "%choice%"=="2" goto build_embedded
if "%choice%"=="3" goto build_both
goto invalid

:build_standard
echo.
echo 构建标准版...
if not exist bin mkdir bin
go build -o bin\webpcompressor.exe cmd\webpcompressor\main.go
if %ERRORLEVEL% EQU 0 (
    echo 标准版构建完成: bin\webpcompressor.exe
) else (
    echo 标准版构建失败
)
goto end

:build_embedded
echo.
echo 构建嵌入版...
if not exist bin mkdir bin
go build -o bin\webptools.exe cmd\embedded\main.go
if %ERRORLEVEL% EQU 0 (
    echo 嵌入版构建完成: bin\webptools.exe
    echo 已嵌入12个WebP工具
) else (
    echo 嵌入版构建失败
)
goto end

:build_both
echo.
echo 构建标准版...
if not exist bin mkdir bin
go build -o bin\webpcompressor.exe cmd\webpcompressor\main.go
if %ERRORLEVEL% EQU 0 (
    echo 标准版构建完成: bin\webpcompressor.exe
) else (
    echo 标准版构建失败
)

echo.
echo 构建嵌入版...
go build -o bin\webptools.exe cmd\embedded\main.go
if %ERRORLEVEL% EQU 0 (
    echo 嵌入版构建完成: bin\webptools.exe
) else (
    echo 嵌入版构建失败
)
goto end

:invalid
echo 无效选择
goto end

:end
echo.
echo 构建结果:
if exist bin\*.exe (
    dir bin\*.exe /B 2>nul | findstr /R ".*\.exe"
) else (
    echo bin目录中没有找到可执行文件
)
echo.
echo 使用说明:
echo   标准版: bin\webpcompressor.exe input.webp 30 output.webp  
echo   嵌入版: bin\webptools.exe compress input.webp 30 output.webp
echo   帮助:   bin\webptools.exe help
echo.
pause 