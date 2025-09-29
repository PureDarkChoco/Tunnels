package config

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// TunnelConfig SSH 터널 설정
type TunnelConfig struct {
	Name        string `yaml:"name"`
	LocalPort   int    `yaml:"local_port"`
	RemoteHost  string `yaml:"remote_host"`
	RemotePort  int    `yaml:"remote_port"`
	SSHHost     string `yaml:"ssh_host"`
	SSHPort     int    `yaml:"ssh_port"`
	SSHUser     string `yaml:"ssh_user"`
	SSHKeyPath  string `yaml:"ssh_key_path,omitempty"`
	SSHPassword string `yaml:"ssh_password,omitempty"`
	Enabled     bool   `yaml:"enabled"`
}

// Config 전체 설정
type Config struct {
	Tunnels []TunnelConfig `yaml:"tunnels"`
	CheckInterval int      `yaml:"check_interval"` // 초 단위
}

// DefaultConfig 기본 설정 생성
func DefaultConfig() *Config {
	return &Config{
		Tunnels: []TunnelConfig{},
		CheckInterval: 30,
	}
}

// LoadConfig 설정 파일 로드
func LoadConfig(configPath string) (*Config, error) {
	// 파일 존재 여부 및 권한 확인
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("기본 설정 파일 생성 실패: %v", err)
		}
		return config, nil
	}

	if err != nil {
		return nil, fmt.Errorf("설정 파일 접근 실패: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패: %v", err)
	}

	// 기본값 설정
	if config.CheckInterval <= 0 {
		config.CheckInterval = 30
	}

	return &config, nil
}

// SaveConfig 설정 파일 저장
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("설정 마샬링 실패: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %v", err)
	}

	return nil
}

// GetEnabledTunnels 활성화된 터널만 반환
func (c *Config) GetEnabledTunnels() []TunnelConfig {
	var enabled []TunnelConfig
	for _, tunnel := range c.Tunnels {
		if tunnel.Enabled {
			enabled = append(enabled, tunnel)
		}
	}
	return enabled
}

// Validate 터널 설정 유효성 검사
func (t *TunnelConfig) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("터널 이름이 필요합니다")
	}
	if t.LocalPort <= 0 || t.LocalPort > 65535 {
		return fmt.Errorf("유효하지 않은 로컬 포트: %d", t.LocalPort)
	}
	if t.RemoteHost == "" {
		return fmt.Errorf("원격 호스트가 필요합니다")
	}
	if t.RemotePort <= 0 || t.RemotePort > 65535 {
		return fmt.Errorf("유효하지 않은 원격 포트: %d", t.RemotePort)
	}
	if t.SSHHost == "" {
		return fmt.Errorf("SSH 호스트가 필요합니다")
	}
	if t.SSHPort <= 0 || t.SSHPort > 65535 {
		return fmt.Errorf("유효하지 않은 SSH 포트: %d", t.SSHPort)
	}
	if t.SSHUser == "" {
		return fmt.Errorf("SSH 사용자명이 필요합니다")
	}
	if t.SSHKeyPath == "" && t.SSHPassword == "" {
		return fmt.Errorf("SSH 키 파일 또는 패스워드가 필요합니다")
	}
	return nil
}

// GetCheckIntervalDuration 체크 간격을 Duration으로 반환
func (c *Config) GetCheckIntervalDuration() time.Duration {
	return time.Duration(c.CheckInterval) * time.Second
}

// CheckKeyFilePermissions 키 파일 권한 확인
func (t *TunnelConfig) CheckKeyFilePermissions() error {
	// 키 파일이 설정되지 않은 경우 (패스워드 인증 사용) 패스
	if t.SSHKeyPath == "" {
		return nil
	}

	// 파일 존재 여부 확인
	info, err := os.Stat(t.SSHKeyPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("SSH 키 파일이 존재하지 않습니다: %s", t.SSHKeyPath)
	}
	if err != nil {
		return fmt.Errorf("SSH 키 파일 접근 실패: %s, 오류: %v", t.SSHKeyPath, err)
	}

	// 파일 권한 확인
	mode := info.Mode()

	// Windows에서는 파일 권한 체크 방식이 다름
	if runtime.GOOS == "windows" {
		// Windows에서는 파일이 읽기 가능한지만 확인
		if mode&0400 == 0 {
			return fmt.Errorf("SSH 키 파일 읽기 권한이 없습니다: %s (권한: %s)", t.SSHKeyPath, mode.String())
		}
	} else {
		// Unix/Linux 시스템에서는 파일 권한이 너무 넓으면 경고
		// SSH는 보안상 키 파일의 권한이 600 (소유자만 읽기/쓰기)이어야 함
		if mode&0077 != 0 {
			return fmt.Errorf("SSH 키 파일 권한이 너무 넓습니다: %s (권한: %s, 권장: 600)", t.SSHKeyPath, mode.String())
		}
	}

	return nil
}

// CheckAllEnabledKeyFilePermissions 모든 활성화된 터널의 키 파일 권한 확인
func (c *Config) CheckAllEnabledKeyFilePermissions() []error {
	var errors []error

	for _, tunnel := range c.GetEnabledTunnels() {
		if err := tunnel.CheckKeyFilePermissions(); err != nil {
			errors = append(errors, fmt.Errorf("터널 '%s': %v", tunnel.Name, err))
		}
	}

	return errors
}
