@echo off
echo Building Tunnels application...

REM 기존 실행 파일이 실행 중이면 종료
taskkill /f /im tunnels.exe >nul 2>&1

REM 의존성 설치
echo Installing dependencies...
go mod tidy

REM 리소스 파일 생성 (아이콘 포함)
echo Generating resource file...
rsrc -ico assets\icons\tunnels.ico -o rsrc.syso

REM 실행 파일 빌드 (Windows GUI 애플리케이션으로 빌드)
echo Building executable...
go build -ldflags "-s -w -H windowsgui" -buildmode=exe -o tunnels.exe

if %ERRORLEVEL% EQU 0 (
    echo Build successful! tunnels.exe created.
    echo.
    echo Usage:
    echo   tunnels.exe [config_file]
    echo.
echo Example:
echo   tunnels.exe tunnels.conf
) else (
    echo Build failed!
    exit /b 1
)
