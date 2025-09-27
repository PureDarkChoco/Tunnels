# Tunnels - SSH 터널 관리자

Windows 시스템 트레이에 상주하면서 SSH 터널을 자동으로 관리하는 Go 애플리케이션입니다.

## 주요 기능

- 🔄 **자동 터널 관리**: 설정 파일 기반으로 SSH 터널을 자동으로 생성하고 관리
- 🔍 **연결 상태 모니터링**: 주기적으로 연결 상태를 확인하고 끊어진 연결을 자동으로 재연결
- 🖥️ **시스템 트레이 통합**: Windows 시스템 트레이에서 간편하게 관리
- ⚙️ **동적 설정 적용**: 설정 파일 변경 시 실시간으로 반영
- 📊 **상태 표시**: 각 터널의 연결 상태를 시각적으로 확인

## 설치 및 실행

### 전제 조건

- Windows 10/11
- Go 1.21 이상
- Windows OpenSSH 클라이언트 (Windows 10 1809+ 기본 포함)

### 빌드

```bash
# 의존성 설치
go mod tidy

# 실행 파일 빌드
go build -o tunnels.exe
```

### 실행

```bash
# 예제 설정 파일을 복사하여 실제 설정 파일 생성
copy tunnels.conf.example tunnels.conf

# 설정 파일 편집 (실제 서버 정보로 수정)
# tunnels.conf

# 기본 설정 파일(tunnels.conf) 사용
./tunnels.exe

# 사용자 정의 설정 파일 사용
./tunnels.exe config.conf
```

## 설정 파일

`tunnels.conf` 파일을 통해 터널 설정을 관리합니다.

### 설정 파일 구조

```yaml
tunnels:
  - name: "터널 이름"           # 터널 식별자 (고유해야 함)
    local_port: 8080           # 로컬 포트
    remote_host: "localhost"   # 원격 호스트
    remote_port: 80            # 원격 포트
    ssh_host: "example.com"    # SSH 서버 호스트
    ssh_port: 22               # SSH 서버 포트
    ssh_user: "username"       # SSH 사용자명
    ssh_key_path: "key.pem"    # SSH 키 파일 경로 (권장)
    # ssh_password: "pass"     # SSH 패스워드 (키 파일이 없을 때)
    enabled: true              # 터널 활성화 여부

check_interval: 30            # 연결 상태 체크 간격 (초)
```

### SSH 키 기반 인증 설정 (권장)

1. SSH 키 생성:
```bash
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```

2. 공개키를 서버에 등록:
```bash
ssh-copy-id user@example.com
```

3. 설정 파일에 개인키 경로 지정:
```yaml
ssh_key_path: "C:\\Users\\YourUser\\.ssh\\id_rsa"
```

## 시스템 트레이 메뉴

### 연결 현황
- 각 터널의 현재 상태를 표시
- ● 연결됨, ⊙ 연결 중, ⊗ 오류, ○ 비활성화
- 터널 클릭 시 해당 터널 재시작

### 메뉴 옵션
- **설정 다시 로드**: 설정 파일을 다시 읽어서 적용
- **설정 파일 열기**: 기본 편집기로 설정 파일 열기
- **모든 터널 재시작**: 모든 활성 터널을 재시작
- **종료**: 애플리케이션 종료

## 문제 해결

### 일반적인 문제

1. **SSH 연결 실패**
   - SSH 서버가 실행 중인지 확인
   - 방화벽 설정 확인
   - SSH 키 권한 확인 (Windows: 600 권한)

2. **포트 충돌**
   - 로컬 포트가 다른 프로그램에서 사용 중인지 확인
   - `netstat -an | findstr :포트번호` 명령으로 확인

3. **권한 문제**
   - 관리자 권한으로 실행 시도
   - SSH 키 파일 권한 확인

### 로그 확인

애플리케이션은 콘솔에 상세한 로그를 출력합니다. 문제 발생 시 로그를 확인하여 원인을 파악할 수 있습니다.

## 보안 고려사항

- SSH 키 기반 인증 사용 권장
- 패스워드 인증 시 보안 위험 고려
- SSH 키 파일 권한을 적절히 설정 (600)
- 신뢰할 수 있는 서버에만 연결

## 라이선스

MIT License

## 기여

버그 리포트나 기능 제안은 GitHub Issues를 통해 해주세요.
