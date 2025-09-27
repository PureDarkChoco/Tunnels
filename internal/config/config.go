package config

import (
	"fmt"
	"os"
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
	// 파일이 존재하지 않으면 기본 설정으로 생성
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("기본 설정 파일 생성 실패: %v", err)
		}
		return config, nil
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
