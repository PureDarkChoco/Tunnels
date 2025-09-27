# Tunnels SSH Tunnel Manager - 배포 가이드

## 📦 배포 파일 구성

```
deploy/
├── tunnels.exe          # 메인 실행 파일 (약 4.8MB, 아이콘 포함)
├── tunnels.conf         # 설정 파일 (예시)
└── README_DEPLOY.md     # 이 파일
```

**🎯 아이콘 내장**: 모든 아이콘 파일이 실행파일에 포함되어 별도 assets 폴더 불필요!
**🚫 콘솔 창 숨김**: 실행 시 콘솔 창이 나타나지 않습니다!

## 🚀 설치 및 실행

### 1. 필수 요구사항
- **Windows 10/11** (64비트 권장)
- **OpenSSH 클라이언트** (Windows 10 1809+ 기본 포함)
- **SSH 키 파일** (설정에서 지정한 경로에 위치)

### 2. 설치 방법
1. `deploy` 폴더의 파일들을 원하는 위치에 복사
2. **예제 설정 파일을 복사하여 실제 설정 파일 생성**:
   ```bash
   copy tunnels.conf.example tunnels.conf
   ```
3. `tunnels.conf` 파일을 편집하여 SSH 연결 정보 설정
4. SSH 키 파일을 지정된 경로에 배치
5. **아이콘 파일 불필요**: 모든 아이콘이 실행파일에 내장됨

### 3. 실행 방법
```bash
# 기본 설정 파일 사용
tunnels.exe

# 특정 설정 파일 사용
tunnels.exe my-config.conf
```

## ⚙️ 설정 파일 예시

```yaml
check_interval: 15

tunnels:
  dev-web:
    local_port: 30
    remote_host: 10.0.10.181
    remote_port: 22
    ssh_host: 54.180.41.74
    ssh_port: 22
    ssh_user: ec2-user
    ssh_key_path: "D:/ssh/hivecity_dev_ed25519.pem"

  kakao:
    local_port: 34
    remote_host: 10.0.2.179
    remote_port: 22
    ssh_host: 54.180.41.74
    ssh_port: 22
    ssh_user: ec2-user
    ssh_key_path: "D:/ssh/hivecity_dev_ed25519.pem"
```

## 🔧 SSH 키 파일 설정

### Windows에서 SSH 키 권한 설정
```powershell
# 권한 제거
icacls "D:\ssh\hivecity_dev_ed25519.pem" /remove "NT AUTHORITY\Authenticated Users"

# 현재 사용자에게 권한 부여
icacls "D:\ssh\hivecity_dev_ed25519.pem" /grant "%USERNAME%:F"
```

## 📋 시스템 요구사항

### 최소 요구사항
- Windows 10 (버전 1809 이상)
- RAM: 50MB
- 디스크: 10MB
- 네트워크 연결

### 권장 사항
- Windows 11
- RAM: 100MB
- SSD 저장소
- 안정적인 인터넷 연결

## 🎯 기능

- ✅ **자동 SSH 터널링**: 설정된 터널 자동 연결
- ✅ **자동 재연결**: 연결 끊김 시 자동 복구
- ✅ **시스템 트레이**: 백그라운드 실행
- ✅ **실시간 상태**: 연결 상태 실시간 모니터링
- ✅ **설정 리로드**: 설정 변경 시 자동 적용
- ✅ **로그 기록**: 연결 상태 및 오류 로그

## 🛠️ 문제 해결

### 실행이 안 되는 경우
1. Windows Defender나 백신 프로그램에서 차단 여부 확인
2. 관리자 권한으로 실행 시도
3. `tunnels.log` 파일에서 오류 메시지 확인

### SSH 연결이 안 되는 경우
1. SSH 키 파일 경로 및 권한 확인
2. 네트워크 연결 상태 확인
3. SSH 서버 접근 가능 여부 확인

### 아이콘이 안 보이는 경우
1. 시스템 트레이 숨김 아이콘 확인 (^)
2. Windows 설정에서 트레이 아이콘 활성화
3. 애플리케이션 재시작

## 📞 지원

문제가 발생하면 `tunnels.log` 파일의 내용을 확인하여 오류 원인을 파악할 수 있습니다.

---

**버전**: v0.1
**빌드**: 정적 링크 (별도 설치 불필요)
**의존성**: Windows 기본 OpenSSH 클라이언트만 필요
**라이센스**: MIT License (자세한 내용은 LICENSE 파일 참조)
