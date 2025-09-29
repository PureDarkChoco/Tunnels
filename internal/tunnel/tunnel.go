package tunnel

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"tunnels/internal/config"
)

// Status 터널 상태
type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusError        Status = "error"
)

// Tunnel SSH 터널 인스턴스
type Tunnel struct {
	config       config.TunnelConfig
	status       Status
	process      *exec.Cmd
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	lastError    string
	lastCheck    time.Time
	retryCount   int    // 연속 실패 횟수
	maxRetries   int    // 최대 재시도 횟수
	lastSuccess  time.Time // 마지막 성공 시간
}

// NewTunnel 새 터널 인스턴스 생성
func NewTunnel(config config.TunnelConfig) *Tunnel {
	ctx, cancel := context.WithCancel(context.Background())
	return &Tunnel{
		config:     config,
		status:     StatusDisconnected,
		ctx:        ctx,
		cancel:     cancel,
		retryCount: 0,
		maxRetries: 3, // 최대 3번까지 재시도
	}
}

// Start 터널 시작
func (t *Tunnel) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == StatusConnected || t.status == StatusConnecting {
		return nil
	}

	t.status = StatusConnecting
	t.lastError = ""

	// SSH 명령어 구성
	cmd, err := t.buildSSHCommand()
	if err != nil {
		t.status = StatusError
		t.lastError = err.Error()
		return err
	}

	// 프로세스 시작
	t.process = exec.CommandContext(t.ctx, cmd[0], cmd[1:]...)

	// SSH 프로세스의 출력을 /dev/null로 리다이렉트 (Windows에서는 NUL)
	t.process.Stdout = nil
	t.process.Stderr = nil

	// Windows에서 콘솔 창 숨기기
	if runtime.GOOS == "windows" {
		t.process.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}
	}

	if err := t.process.Start(); err != nil {
		t.status = StatusError
		t.lastError = fmt.Sprintf("프로세스 시작 실패: %v", err)
		return err
	}

		// SSH 프로세스 모니터링은 제거 (연결 상태만 체크)

	// 연결 상태 모니터링은 Manager에서 처리

	// 초기 상태는 연결 중으로 설정 (실제 연결 확인 후 변경됨)
	t.status = StatusConnecting
	log.Printf("터널 '%s' SSH 프로세스 시작됨: %s", t.config.Name, t.GetConnectionString())
	return nil
}

// Stop 터널 중지
func (t *Tunnel) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == StatusDisconnected {
		return nil
	}

	t.cancel()

	if t.process != nil && t.process.Process != nil {
		// Windows에서는 직접 Kill 사용 (SIGTERM 미지원)
		if runtime.GOOS == "windows" {
			if err := t.process.Process.Kill(); err != nil {
				log.Printf("터널 '%s' 프로세스 강제 종료 실패: %v", t.config.Name, err)
			}
		} else {
			// Linux/macOS에서는 SIGTERM으로 정상 종료 시도
			if err := t.process.Process.Signal(os.Interrupt); err != nil {
				log.Printf("터널 '%s' 프로세스 SIGTERM 실패: %v", t.config.Name, err)
			}

			// 잠시 대기 후 강제 종료
			time.Sleep(2 * time.Second)

			// ProcessState가 nil이 아닌지 확인 후 Exited() 호출
			if t.process.ProcessState != nil && !t.process.ProcessState.Exited() {
				if err := t.process.Process.Kill(); err != nil {
					log.Printf("터널 '%s' 프로세스 강제 종료 실패: %v", t.config.Name, err)
				}
			}
		}

		// 프로세스 완전 종료 대기
		time.Sleep(1 * time.Second)
	}

	t.status = StatusDisconnected
	t.lastError = ""
	log.Printf("터널 '%s' 중지됨", t.config.Name)
	return nil
}

// Restart 터널 재시작
func (t *Tunnel) Restart() error {
	t.mu.Lock()

	// 재시도 횟수가 최대값에 도달한 경우 재시작하지 않음
	if t.retryCount >= t.maxRetries {
		t.mu.Unlock()
		log.Printf("터널 '%s' 최대 재시도 횟수(%d) 초과로 재시작 중단", t.config.Name, t.maxRetries)
		return fmt.Errorf("최대 재시도 횟수(%d) 초과", t.maxRetries)
	}

	// 기존 프로세스 중지
	if t.process != nil && t.process.Process != nil {
		t.process.Process.Kill()
	}

	// 새 context 생성 (기존 context가 취소되었을 수 있음)
	t.ctx, t.cancel = context.WithCancel(context.Background())

	t.mu.Unlock()

	time.Sleep(1 * time.Second) // 잠시 대기
	return t.Start()
}

// GetStatus 터널 상태 반환
func (t *Tunnel) GetStatus() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetConfig 설정 반환
func (t *Tunnel) GetConfig() config.TunnelConfig {
	return t.config
}

// GetLastError 마지막 에러 반환
func (t *Tunnel) GetLastError() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastError
}

// GetLastCheck 마지막 체크 시간 반환
func (t *Tunnel) GetLastCheck() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastCheck
}

// buildSSHCommand SSH 명령어 구성
func (t *Tunnel) buildSSHCommand() ([]string, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("이 프로그램은 Windows에서만 동작합니다")
	}

	// Windows OpenSSH 클라이언트 사용
	cmd := []string{"ssh"}

	// 포트 설정
	if t.config.SSHPort != 22 {
		cmd = append(cmd, "-p", strconv.Itoa(t.config.SSHPort))
	}

	// 로컬 포트 포워딩 설정
	localForward := fmt.Sprintf("%d:%s:%d", t.config.LocalPort, t.config.RemoteHost, t.config.RemotePort)
	cmd = append(cmd, "-L", localForward)

	// 키 파일 설정
	if t.config.SSHKeyPath != "" {
		cmd = append(cmd, "-i", t.config.SSHKeyPath)
	}

	// 패스워드 인증 사용 시
	if t.config.SSHPassword != "" {
		// Windows에서는 sshpass가 기본 제공되지 않으므로 키 기반 인증 권장
		log.Printf("경고: 패스워드 인증은 Windows OpenSSH에서 제한적입니다. 키 기반 인증을 권장합니다.")
	}

	// 연결 타임아웃 설정
	cmd = append(cmd, "-o", "ConnectTimeout=10")
	cmd = append(cmd, "-o", "ServerAliveInterval=20")
	cmd = append(cmd, "-o", "ServerAliveCountMax=3")

	// 호스트 키 확인 비활성화 (개발용)
	cmd = append(cmd, "-o", "StrictHostKeyChecking=no")
	cmd = append(cmd, "-o", "UserKnownHostsFile=NUL")

	// 백그라운드 실행을 위한 옵션
	cmd = append(cmd, "-N")

	// 사용자@호스트
	cmd = append(cmd, fmt.Sprintf("%s@%s", t.config.SSHUser, t.config.SSHHost))

	return cmd, nil
}



// CheckConnection 연결 상태 확인
func (t *Tunnel) CheckConnection() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lastCheck = time.Now()

	// 이미 오류 상태이고 재시도 횟수가 최대값에 도달한 경우 연결 확인하지 않음
	if t.status == StatusError && t.retryCount >= t.maxRetries {
		return
	}

	// 로컬 포트가 열려있는지 확인
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", t.config.LocalPort), 5*time.Second)
	if err != nil {
		if t.status == StatusConnected || t.status == StatusConnecting {
			t.retryCount++
			t.status = StatusError
			t.lastError = fmt.Sprintf("로컬 포트 %d 연결 실패: %v (재시도 %d/%d)", t.config.LocalPort, err, t.retryCount, t.maxRetries)

			if t.retryCount >= t.maxRetries {
				log.Printf("터널 '%s' 최대 재시도 횟수(%d) 초과 - 자동 재시작 중단", t.config.Name, t.maxRetries)
				t.lastError = fmt.Sprintf("최대 재시도 횟수(%d) 초과: %v", t.maxRetries, err)
			} else {
				log.Printf("터널 '%s' 로컬 포트 연결 실패: %v - 자동 재시작 시도 (%d/%d)", t.config.Name, err, t.retryCount, t.maxRetries)

				// 자동 재시작 시도 (뮤텍스 해제 후)
				go func() {
					time.Sleep(2 * time.Second) // 잠시 대기 후 재시작
					if err := t.Restart(); err != nil {
						log.Printf("터널 '%s' 자동 재시작 실패: %v", t.config.Name, err)
					}
				}()
			}
		}
		return
	}
	conn.Close()

	// 연결 성공 시 상태 업데이트 및 재시도 횟수 리셋
	if t.status == StatusConnecting {
		t.status = StatusConnected
		t.lastError = ""
		t.retryCount = 0 // 재시도 횟수 리셋
		t.lastSuccess = time.Now()
		log.Printf("터널 '%s' 연결 성공", t.config.Name)
	} else if t.status == StatusError {
		t.status = StatusConnected
		t.lastError = ""
		t.retryCount = 0 // 재시도 횟수 리셋
		t.lastSuccess = time.Now()
		log.Printf("터널 '%s' 연결 복구됨", t.config.Name)
	}
}

// GetConnectionString 연결 문자열 반환
func (t *Tunnel) GetConnectionString() string {
	return fmt.Sprintf("127.0.0.1:%d -> %s:%d (via %s@%s:%d)",
		t.config.LocalPort,
		t.config.RemoteHost,
		t.config.RemotePort,
		t.config.SSHUser,
		t.config.SSHHost,
		t.config.SSHPort)
}

// IsHealthy 터널이 정상 상태인지 확인
func (t *Tunnel) IsHealthy() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status == StatusConnected
}

// UpdateConfig 설정 업데이트
func (t *Tunnel) UpdateConfig(newConfig config.TunnelConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 설정이 변경되었으면 재시작
	if t.config != newConfig {
		t.config = newConfig
		if t.status == StatusConnected {
			go t.Restart()
		}
	}
}

// SetErrorStatus 오류 상태 설정 (권한 문제 등으로 연결할 수 없는 경우)
func (t *Tunnel) SetErrorStatus(errorMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.status = StatusError
	t.lastError = errorMsg
	t.retryCount = t.maxRetries // 재시도 횟수를 최대값으로 설정하여 재시도 방지
	log.Printf("터널 '%s' 오류 상태 설정: %s", t.config.Name, errorMsg)
}