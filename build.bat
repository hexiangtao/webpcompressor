@echo off
chcp 65001 > nul
echo 🎨 WebP Multi-Tool Builder
echo =======================================

echo.
echo 📦 Build Options:
echo   1. Standard (External tools required) - webpcompressor.exe
echo   2. Embedded (All tools included) - webptools.exe  
echo   3. Build both versions
echo.

set /p choice="Choose build option (1/2/3): "

if "%choice%"=="1" goto build_standard
if "%choice%"=="2" goto build_embedded
if "%choice%"=="3" goto build_both
goto invalid

:build_standard
echo.
echo 🔧 Building Standard Version...
go build -o webpcompressor.exe cmd/webpcompressor/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Standard version built: webpcompressor.exe
) else (
    echo ❌ Standard version build failed
)
goto end

:build_embedded
echo.
echo 🔧 Building Embedded Version...
go build -o webptools.exe cmd/embedded/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Embedded version built: webptools.exe
    echo 📁 Embedded 12 WebP tools
) else (
    echo ❌ Embedded version build failed
)
goto end

:build_both
echo.
echo 🔧 Building Standard Version...
go build -o webpcompressor.exe cmd/webpcompressor/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Standard version built: webpcompressor.exe
) else (
    echo ❌ Standard version build failed
)

echo.
echo 🔧 Building Embedded Version...
go build -o webptools.exe cmd/embedded/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Embedded version built: webptools.exe
) else (
    echo ❌ Embedded version build failed
)
goto end

:invalid
echo ❌ Invalid choice
goto end

:end
echo.
echo 📊 Build Results:
dir *.exe 2>nul | findstr /R "webp.*\.exe"
echo.
echo 💡 Usage:
echo   Standard: webpcompressor.exe input.webp 30 output.webp  
echo   Embedded: webptools.exe compress input.webp 30 output.webp
echo   Help:     webptools.exe help
echo.
pause 