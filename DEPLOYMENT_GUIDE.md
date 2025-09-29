# Tunnels SSH 터널 관리자 - 배포 가이드

## 📁 배포 파일 목록

다른 PC에 복사해야 할 파일들:
- `tunnels.exe` - 메인 실행 파일 (4.7MB)
- `tunnels.conf` - 설정 파일
- `fix_ssh_permissions.bat` - SSH 권한 수정 도구 (상세 버전)
- `fix_ssh_permissions_simple.bat` - SSH 권한 수정 도구 (간단 버전)
- `build.bat` - 빌드 스크립트 (개발용)

## 🚀 설치 및 실행 방법

### 1단계: 파일 복사
1. 위 파일들을 다른 PC의 적절한 폴더에 복사
2. SSH 키 파일들을 `C:\Users\YourName\.ssh\` 경로에 배치

### 2단계: SSH 키 권한 수정
**관리자 권한으로** 다음 중 하나를 실행:
- `fix_ssh_permissions.bat` (상세 버전 - 권장)
- `fix_ssh_permissions_simple.bat` (간단 버전)

### 3단계: 설정 파일 수정
`tunnels.conf` 파일에서 다음을 수정:
1. **SSH 서버 정보**: 실제 서버 주소, 포트, 사용자명으로 변경
2. **SSH 키 경로**: 실제 SSH 키 파일 경로로 변경
```yaml
ssh_host: "your-server.example.com"  # 실제 서버 주소
ssh_user: "your-username"            # 실제 사용자명
ssh_key_path: "C:\\Users\\YourName\\.ssh\\your_key.pem"  # 실제 키 파일 경로
```

### 4단계: 애플리케이션 실행
```cmd
tunnels.exe
```

## 🔧 문제 해결

### SSH 연결 실패 시
1. SSH 키 파일 권한 확인:
   ```cmd
   icacls "C:\Users\%USERNAME%\.ssh\your_key.pem"
   ```

2. SSH 연결 직접 테스트:
   ```cmd
   ssh -i "C:\Users\%USERNAME%\.ssh\your_key.pem" -o ConnectTimeout=10 -o StrictHostKeyChecking=no user@server "echo 'test'"
   ```

3. Windows OpenSSH 클라이언트 설치 확인:
   ```cmd
   ssh -V
   ```

### 터널 연결 실패 시
1. 방화벽 설정 확인
2. 네트워크 연결 상태 확인
3. AWS 서버 접근 가능 여부 확인

## 📋 시스템 요구사항

- **운영체제**: Windows 10/11
- **OpenSSH 클라이언트**: Windows 10 1809+ 기본 포함
- **권한**: SSH 키 파일 수정을 위한 관리자 권한
- **네트워크**: 인터넷 연결 및 AWS 서버 접근 가능

## 🔄 업데이트 방법

1. 새로운 `tunnels.exe` 파일로 교체
2. 설정 파일(`tunnels.conf`) 필요시 수정
3. SSH 키 권한 재설정 (필요시)

## 📞 지원

문제가 발생하면 다음을 확인하세요:
1. `tunnels.log` 파일의 오류 메시지
2. SSH 키 파일 권한
3. 네트워크 연결 상태
4. 방화벽 설정
