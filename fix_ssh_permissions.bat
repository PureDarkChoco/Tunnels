@echo off
echo ========================================
echo SSH 키 파일 권한 수정 도구
echo ========================================
echo.

REM 관리자 권한 확인
net session >nul 2>&1
if %errorLevel% == 0 (
    echo [INFO] 관리자 권한으로 실행 중입니다.
) else (
    echo [ERROR] 이 스크립트는 관리자 권한으로 실행해야 합니다.
    echo [INFO] 마우스 우클릭 후 "관리자 권한으로 실행"을 선택하세요.
    pause
    exit /b 1
)

echo.
echo [INFO] SSH 키 파일 권한을 수정하는 중...

REM SSH 키 디렉토리 존재 확인
if not exist "C:\Users\%USERNAME%\.ssh" (
    echo [INFO] C:\Users\%USERNAME%\.ssh 디렉토리를 생성합니다.
    mkdir "C:\Users\%USERNAME%\.ssh" 2>nul
)

REM 각 SSH 키 파일의 권한 수정
echo.
echo [1/3] dev_key.pem 권한 수정 중...
if exist "C:\Users\%USERNAME%\.ssh\dev_key.pem" (
    icacls "C:\Users\%USERNAME%\.ssh\dev_key.pem" /reset >nul 2>&1
    icacls "C:\Users\%USERNAME%\.ssh\dev_key.pem" /inheritance:r /grant:r "%USERNAME%:F" >nul 2>&1
    if %errorLevel% == 0 (
        echo [OK] dev_key.pem 권한 수정 완료
    ) else (
        echo [ERROR] dev_key.pem 권한 수정 실패
    )
) else (
    echo [WARNING] dev_key.pem 파일이 존재하지 않습니다. tunnels.conf에서 경로를 확인하세요.
)

echo.
echo [2/3] prod_key.pem 권한 수정 중...
if exist "C:\Users\%USERNAME%\.ssh\prod_key.pem" (
    icacls "C:\Users\%USERNAME%\.ssh\prod_key.pem" /reset >nul 2>&1
    icacls "C:\Users\%USERNAME%\.ssh\prod_key.pem" /inheritance:r /grant:r "%USERNAME%:F" >nul 2>&1
    if %errorLevel% == 0 (
        echo [OK] prod_key.pem 권한 수정 완료
    ) else (
        echo [ERROR] prod_key.pem 권한 수정 실패
    )
) else (
    echo [WARNING] prod_key.pem 파일이 존재하지 않습니다. tunnels.conf에서 경로를 확인하세요.
)

echo.
echo [3/3] 추가 SSH 키 파일 권한 수정 중...
REM tunnels.conf에서 설정한 다른 SSH 키 파일들이 있다면 여기에 추가하세요.
REM 예: id_rsa, id_ed25519, custom_key.pem 등
if exist "C:\Users\%USERNAME%\.ssh\id_rsa" (
    echo [INFO] id_rsa 권한 수정 중...
    icacls "C:\Users\%USERNAME%\.ssh\id_rsa" /reset >nul 2>&1
    icacls "C:\Users\%USERNAME%\.ssh\id_rsa" /inheritance:r /grant:r "%USERNAME%:F" >nul 2>&1
    if %errorLevel% == 0 (
        echo [OK] id_rsa 권한 수정 완료
    ) else (
        echo [ERROR] id_rsa 권한 수정 실패
    )
)

if exist "C:\Users\%USERNAME%\.ssh\id_ed25519" (
    echo [INFO] id_ed25519 권한 수정 중...
    icacls "C:\Users\%USERNAME%\.ssh\id_ed25519" /reset >nul 2>&1
    icacls "C:\Users\%USERNAME%\.ssh\id_ed25519" /inheritance:r /grant:r "%USERNAME%:F" >nul 2>&1
    if %errorLevel% == 0 (
        echo [OK] id_ed25519 권한 수정 완료
    ) else (
        echo [ERROR] id_ed25519 권한 수정 실패
    )
)

echo.
echo ========================================
echo SSH 연결 테스트 시작
echo ========================================
echo.

REM SSH 연결 테스트 (키 파일이 존재하는 경우에만)
echo [TEST] SSH 연결 테스트 중...
echo [INFO] tunnels.conf에서 설정한 SSH 서버 정보로 연결 테스트를 수행합니다.
echo [INFO] SSH 키 파일이 올바른 위치에 있고 권한이 설정되었는지 확인하세요.

REM 사용자가 tunnels.conf에서 실제 서버 정보를 설정한 후 테스트하도록 안내
echo [INFO] SSH 연결 테스트를 하려면:
echo [INFO] 1. tunnels.conf에서 실제 서버 정보를 설정하세요
echo [INFO] 2. SSH 키 파일을 올바른 경로에 배치하세요
echo [INFO] 3. 다음 명령어로 직접 테스트하세요:
echo [INFO]    ssh -i "C:\Users\%USERNAME%\.ssh\dev_key.pem" user@server "echo test"

echo.
echo ========================================
echo 권한 수정 완료!
echo ========================================
echo.
echo [INFO] 이제 Tunnels 애플리케이션을 실행할 수 있습니다.
echo [INFO] tunnels.exe 파일을 실행하세요.
echo.
pause
