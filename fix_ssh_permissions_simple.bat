@echo off
echo SSH 키 파일 권한 수정 중...

REM 관리자 권한 확인
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 관리자 권한으로 실행하세요.
    pause
    exit /b 1
)

REM SSH 디렉토리 생성
if not exist "C:\Users\%USERNAME%\.ssh" mkdir "C:\Users\%USERNAME%\.ssh" 2>nul

REM 모든 SSH 키 파일 권한 수정 (일반적인 SSH 키 파일들)
for %%f in (
    "C:\Users\%USERNAME%\.ssh\dev_key.pem"
    "C:\Users\%USERNAME%\.ssh\prod_key.pem"
    "C:\Users\%USERNAME%\.ssh\id_rsa"
    "C:\Users\%USERNAME%\.ssh\id_ed25519"
    "C:\Users\%USERNAME%\.ssh\*.pem"
    "C:\Users\%USERNAME%\.ssh\*.ppk"
) do (
    if exist %%f (
        echo 수정 중: %%f
        icacls %%f /reset >nul 2>&1
        icacls %%f /inheritance:r /grant:r "%USERNAME%:F" >nul 2>&1
    )
)

echo 완료! Tunnels를 실행하세요.
pause
