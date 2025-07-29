@echo off
chcp 65001 > nul
echo 🎨 WebP Multi-Tool Builder
echo =======================================

echo.
echo 📦 Build Options:
echo   1. Standard (External tools required) - bin/webpcompressor.exe
echo   2. Embedded (All tools included) - bin/webptools.exe  
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
if not exist bin mkdir bin
go build -o bin/webpcompressor.exe cmd/webpcompressor/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Standard version built: bin/webpcompressor.exe
) else (
    echo ❌ Standard version build failed
)
goto end

:build_embedded
echo.
echo 🔧 Building Embedded Version...
if not exist bin mkdir bin
go build -o bin/webptools.exe cmd/embedded/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Embedded version built: bin/webptools.exe
    echo 📁 Embedded 12 WebP tools
) else (
    echo ❌ Embedded version build failed
)
goto end

:build_both
echo.
echo 🔧 Building Standard Version...
if not exist bin mkdir bin
go build -o bin/webpcompressor.exe cmd/webpcompressor/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Standard version built: bin/webpcompressor.exe
) else (
    echo ❌ Standard version build failed
)

echo.
echo 🔧 Building Embedded Version...
go build -o bin/webptools.exe cmd/embedded/main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Embedded version built: bin/webptools.exe
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
if exist bin\*.exe (
    dir bin\*.exe /B 2>nul | findstr /R ".*\.exe"
) else (
    echo No executables found in bin/
)
echo.
echo 💡 Usage:
echo   Standard: bin\webpcompressor.exe input.webp 30 output.webp  
echo   Embedded: bin\webptools.exe compress input.webp 30 output.webp
echo   Help:     bin\webptools.exe help
echo.
pause 