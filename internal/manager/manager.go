package manager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"tunnels/internal/config"
	"tunnels/internal/tunnel"
)

// Manager 터널 매니저
type Manager struct {
	tunnels    map[string]*tunnel.Tunnel
	tunnelOrder []string  // 터널 순서를 유지하기 위한 슬라이스
	config     *config.Config
	configPath string
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewManager 새 매니저 생성
func NewManager(configPath string) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		tunnels:     make(map[string]*tunnel.Tunnel),
		tunnelOrder: make([]string, 0),
		configPath:  configPath,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// LoadConfig 설정 로드 및 터널 업데이트
func (m *Manager) LoadConfig() error {
	cfg, err := config.LoadConfig(m.configPath)
	if err != nil {
		return fmt.Errorf("설정 로드 실패: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 기존 터널들 중지
	for name, t := range m.tunnels {
		t.Stop()
		delete(m.tunnels, name)
	}

	// 터널 순서 초기화
	m.tunnelOrder = make([]string, 0)

	// 새 context 생성 (기존 context가 취소되었을 수 있음)
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// 설정 업데이트
	m.config = cfg

	// 새 터널들 생성 및 시작
	enabledTunnels := cfg.GetEnabledTunnels()
	successCount := 0

	for _, tunnelConfig := range enabledTunnels {
		if err := tunnelConfig.Validate(); err != nil {
			log.Printf("터널 '%s' 설정 오류: %v", tunnelConfig.Name, err)
			continue
		}

		t := tunnel.NewTunnel(tunnelConfig)
		m.tunnels[tunnelConfig.Name] = t
		// 터널 순서 저장 (설정 파일 순서 유지)
		m.tunnelOrder = append(m.tunnelOrder, tunnelConfig.Name)

		// 비동기로 터널 시작
		go func(t *tunnel.Tunnel) {
			if err := t.Start(); err != nil {
				log.Printf("터널 '%s' 시작 실패: %v", t.GetConfig().Name, err)
			} else {
				successCount++
			}
		}(t)
	}

	// 잠시 대기 후 성공한 터널 수 로그 및 즉시 상태 확인
	go func() {
		time.Sleep(1 * time.Second)
		log.Printf("설정 로드 완료: %d개 터널 활성화", successCount)

		// 최초 상태 확인 (즉시 업데이트)
		m.checkAndReconnect()
	}()

	return nil
}

// StartAll 모든 터널 시작
func (m *Manager) StartAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []string
	for name, t := range m.tunnels {
		if err := t.Start(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("일부 터널 시작 실패: %s", fmt.Sprintf("%v", errors))
	}

	return nil
}

// StopAll 모든 터널 중지
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 모든 터널을 병렬로 중지
	var wg sync.WaitGroup
	for name, t := range m.tunnels {
		wg.Add(1)
		go func(tunnelName string, tunnel *tunnel.Tunnel) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("터널 '%s' 중지 중 panic 발생: %v", tunnelName, r)
				}
			}()

			if err := tunnel.Stop(); err != nil {
				log.Printf("터널 '%s' 중지 실패: %v", tunnelName, err)
			}
		}(name, t)
	}

	// 모든 터널 중지 완료 대기
	wg.Wait()

	// 매니저 컨텍스트 취소
	m.cancel()
	return nil
}

// RestartAll 모든 터널 재시작
func (m *Manager) RestartAll() error {
	// LoadConfig에서 이미 모든 터널을 재시작하므로
	// 여기서는 추가 작업이 필요 없음
	log.Println("모든 터널 재시작 완료")
	return nil
}

// GetTunnelStatuses 모든 터널 상태 반환 (순서 보장)
func (m *Manager) GetTunnelStatuses() []TunnelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]TunnelStatus, 0, len(m.tunnels))

	// 설정 파일 순서대로 터널 상태 반환
	for _, name := range m.tunnelOrder {
		if t, exists := m.tunnels[name]; exists {
			statuses = append(statuses, TunnelStatus{
				Name:       name,
				Status:     t.GetStatus(),
				Config:     t.GetConfig(),
				LastError:  t.GetLastError(),
				LastCheck:  t.GetLastCheck(),
				Connection: t.GetConnectionString(),
			})
		}
	}
	return statuses
}

// StartMonitoring 연결 상태 모니터링 시작
func (m *Manager) StartMonitoring() {
	go m.monitorLoop()
}

// monitorLoop 모니터링 루프
func (m *Manager) monitorLoop() {
	ticker := time.NewTicker(m.config.GetCheckIntervalDuration())
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAndReconnect()
		}
	}
}

// checkAndReconnect 연결 상태 확인 및 재연결
func (m *Manager) checkAndReconnect() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, t := range m.tunnels {
		// 연결 상태 확인
		t.CheckConnection()
	}
}

// GetConfigPath 설정 파일 경로 반환
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// GetConfig 현재 설정 반환
func (m *Manager) GetConfig() *config.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// TunnelStatus 터널 상태 정보
type TunnelStatus struct {
	Name       string
	Status     tunnel.Status
	Config     config.TunnelConfig
	LastError  string
	LastCheck  time.Time
	Connection string
}

// GetHealthyCount 정상 동작 중인 터널 수 반환
func (m *Manager) GetHealthyCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, t := range m.tunnels {
		if t.IsHealthy() {
			count++
		}
	}
	return count
}

// GetTotalCount 전체 터널 수 반환
func (m *Manager) GetTotalCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tunnels)
}
