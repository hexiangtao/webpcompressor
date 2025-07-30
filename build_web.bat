@echo off
chcp 65001 >nul

echo Building WebP Compressor Web Service...

:: Set variables
set APP_NAME=webpserver
set VERSION=2.0.0
set BUILD_TIME=%date% %time%
set OUTPUT_DIR=bin
set WEB_DIR=web

:: Create output directory
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

:: Clean previous build
del /f /q %OUTPUT_DIR%\%APP_NAME%.exe 2>nul

:: Set Go build parameters
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

:: Build Web Server
echo Building web server...
go build -ldflags "-X main.Version=%VERSION% -X 'main.BuildTime=%BUILD_TIME%' -s -w" -o %OUTPUT_DIR%\%APP_NAME%.exe cmd/webserver/main.go

if %errorlevel% neq 0 (
    echo Build failed
    exit /b 1
)

:: Copy Web resources
echo Copying web resources...
if not exist %OUTPUT_DIR%\web mkdir %OUTPUT_DIR%\web
if not exist %OUTPUT_DIR%\web\static mkdir %OUTPUT_DIR%\web\static

copy %WEB_DIR%\index.html %OUTPUT_DIR%\web\index.html >nul
copy %WEB_DIR%\static\*.* %OUTPUT_DIR%\web\static\ >nul 2>nul

:: Create default directories
if not exist %OUTPUT_DIR%\uploads mkdir %OUTPUT_DIR%\uploads
if not exist %OUTPUT_DIR%\outputs mkdir %OUTPUT_DIR%\outputs

:: Create sample configuration
echo Creating configuration sample...
echo # WebP Compressor Web Service Configuration > %OUTPUT_DIR%\config.env
echo. >> %OUTPUT_DIR%\config.env
echo # Web Service Configuration >> %OUTPUT_DIR%\config.env
echo WEBP_WEB_HOST=0.0.0.0 >> %OUTPUT_DIR%\config.env
echo WEBP_WEB_PORT=8080 >> %OUTPUT_DIR%\config.env
echo WEBP_WEB_MAX_FILE_SIZE=104857600 >> %OUTPUT_DIR%\config.env
echo WEBP_WEB_ENABLE_AUTH=false >> %OUTPUT_DIR%\config.env
echo WEBP_WEB_AUTH_TOKEN= >> %OUTPUT_DIR%\config.env
echo. >> %OUTPUT_DIR%\config.env
echo # Logging Configuration >> %OUTPUT_DIR%\config.env
echo WEBP_LOG_LEVEL=info >> %OUTPUT_DIR%\config.env
echo. >> %OUTPUT_DIR%\config.env
echo # Processing Configuration >> %OUTPUT_DIR%\config.env
echo WEBP_MAX_CONCURRENCY=4 >> %OUTPUT_DIR%\config.env
echo WEBP_ENABLE_PARALLEL=true >> %OUTPUT_DIR%\config.env

:: Create startup script
echo Creating startup script...
echo @echo off > %OUTPUT_DIR%\start_webserver.bat
echo chcp 65001 ^>nul >> %OUTPUT_DIR%\start_webserver.bat
echo echo Starting WebP Compressor Web Service... >> %OUTPUT_DIR%\start_webserver.bat
echo echo. >> %OUTPUT_DIR%\start_webserver.bat
echo echo Access URL: http://localhost:8080 >> %OUTPUT_DIR%\start_webserver.bat
echo echo Health Check: http://localhost:8080/health >> %OUTPUT_DIR%\start_webserver.bat
echo echo API Info: http://localhost:8080/api/v1/info >> %OUTPUT_DIR%\start_webserver.bat
echo echo. >> %OUTPUT_DIR%\start_webserver.bat
echo echo Press Ctrl+C to stop service >> %OUTPUT_DIR%\start_webserver.bat
echo echo. >> %OUTPUT_DIR%\start_webserver.bat
echo %APP_NAME%.exe >> %OUTPUT_DIR%\start_webserver.bat
echo pause >> %OUTPUT_DIR%\start_webserver.bat

:: Show build result
echo.
echo Build completed successfully!
echo.
echo Output directory: %OUTPUT_DIR%\
echo Web server: %APP_NAME%.exe
echo Startup script: start_webserver.bat
echo Configuration sample: config.env
echo.
echo Usage:
echo    cd %OUTPUT_DIR%
echo    start_webserver.bat
echo.
echo    Or run directly:
echo    %APP_NAME%.exe
echo.
echo Access URL: http://localhost:8080
echo. 