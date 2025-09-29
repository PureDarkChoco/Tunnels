package app

import (
	"embed"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"tunnels/internal/manager"
	"tunnels/internal/tunnel"
	"tunnels/internal/version"

	"github.com/getlantern/systray"
)

// TunnelApp 메인 애플리케이션
type TunnelApp struct {
	manager     *manager.Manager
	configPath  string
	statusItems map[string]*systray.MenuItem
	quitCh      chan bool
	iconPath    string
	iconAssets  embed.FS // 아이콘 에셋
}

// NewTunnelApp 새 앱 인스턴스 생성
func NewTunnelApp(configPath string, iconAssets embed.FS) *TunnelApp {
	return &TunnelApp{
		configPath:  configPath,
		statusItems: make(map[string]*systray.MenuItem),
		quitCh:      make(chan bool),
		iconPath:    "disconnected",
		iconAssets:  iconAssets,
	}
}

// OnReady 시스템 트레이 준비 완료 시 호출
func (app *TunnelApp) OnReady() {
	// 아이콘 설정을 시도하되 실패해도 계속 진행
	defer func() {
		if r := recover(); r != nil {
			log.Printf("시스템 트레이 초기화 중 오류 (무시됨): %v", r)
		}
	}()

	// 기본 설정
	systray.SetTitle(version.AppName)
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("초기 툴팁 설정 에러 (무시됨): %v", r)
			}
		}()
		systray.SetTooltip("SSH Tunnel Manager - Click to view menu")
	}()

	// 초기 아이콘 설정 (disconnected 상태로 시작)
	app.setInitialTrayIcon()

	// 트레이 상태 로그 출력
	app.logTrayStatus()

	// 시작 시 기존 SSH 프로세스 정리
	app.CleanupSSHProcesses()

	// 매니저 초기화
	app.manager = manager.NewManager(app.configPath)

	// 설정 로드 및 터널 시작
	if err := app.manager.LoadConfig(); err != nil {
		log.Printf("초기 설정 로드 실패: %v", err)
		systray.SetTooltip("Tunnels - 설정 로드 실패")
	} else {
		// 터널 시작 후 아이콘 업데이트
		time.Sleep(1 * time.Second) // 터널 시작 대기
		app.updateTrayIcon()
	}

	// 모니터링 시작
	app.manager.StartMonitoring()

	// 메뉴 구성
	app.setupMenu()

	// 상태 업데이트 루프 시작
	go app.updateStatusLoop()

	// 아이콘 업데이트 루프 시작 (더 긴 간격)
	go app.updateIconLoop()
}

// OnExit 시스템 트레이 종료 시 호출
func (app *TunnelApp) OnExit() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("OnExit 중 panic 발생: %v", r)
		}
	}()

	log.Println("애플리케이션 종료 중...")

	if app.manager != nil {
		log.Println("모든 터널 중지 중...")

		// 안전하게 터널 중지
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("터널 중지 중 panic 발생: %v", r)
				}
			}()
			app.manager.StopAll()
		}()

		// SSH 프로세스 정리 대기
		time.Sleep(3 * time.Second)

		// 남은 SSH 프로세스 강제 종료
		app.CleanupSSHProcesses()
	}

	// quitCh 안전하게 닫기
	select {
	case <-app.quitCh:
		// 이미 닫혀있음
	default:
		close(app.quitCh)
	}

	log.Println("애플리케이션 종료 완료")
}

// setupMenu 메뉴 구성
func (app *TunnelApp) setupMenu() {
	// 버전 표시
	versionItem := systray.AddMenuItem(version.AppFullName, "")
	versionItem.Disable()

	// 구분선
	systray.AddSeparator()

	// 터널 상태 메뉴 아이템들 생성
	app.createStatusItems()

	// 구분선
	systray.AddSeparator()

	// 설정 다시 로드 및 터널 재시작 (통합)
	reloadAndRestartItem := systray.AddMenuItem("Reload Config", "Reload config file and restart all tunnels")
	go func() {
		for range reloadAndRestartItem.ClickedCh {
			app.reloadConfigAndRestart()
		}
	}()

	// 설정 파일 열기
	openConfigItem := systray.AddMenuItem("Open Config File", "Open config file in editor")
	go func() {
		for range openConfigItem.ClickedCh {
			app.openConfigFile()
		}
	}()

	// 구분선
	systray.AddSeparator()

	// 종료
	quitItem := systray.AddMenuItem("Exit", "Exit application")
	go func() {
		for range quitItem.ClickedCh {
			systray.Quit()
		}
	}()
}

// reloadConfigAndRestart 설정 다시 로드 및 모든 터널 재시작
func (app *TunnelApp) reloadConfigAndRestart() {
	log.Println("설정 다시 로드 및 터널 재시작 중...")

	// 설정 다시 로드 (이미 터널 재시작도 포함됨)
	if err := app.manager.LoadConfig(); err != nil {
		log.Printf("설정 로드 실패: %v", err)
		app.showError("설정 로드 실패", err.Error())
		return
	}

	// 메뉴 업데이트 (재구성하지 않고 기존 항목들만 업데이트)
	app.updateMenuForConfigReload()

	// 아이콘 업데이트
	app.updateTrayIcon()

	log.Println("설정 다시 로드 및 터널 재시작 완료")
}

// openConfigFile 설정 파일 열기
func (app *TunnelApp) openConfigFile() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("notepad.exe", app.configPath)
	case "darwin":
		cmd = exec.Command("open", "-e", app.configPath)
	case "linux":
		cmd = exec.Command("xdg-open", app.configPath)
	default:
		log.Printf("지원하지 않는 OS: %s", runtime.GOOS)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("설정 파일 열기 실패: %v", err)
		app.showError("설정 파일 열기 실패", err.Error())
	}
}

// updateStatusLoop 상태 업데이트 루프
func (app *TunnelApp) updateStatusLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.quitCh:
			return
		case <-ticker.C:
			app.updateStatus()
		}
	}
}

// updateIconLoop 아이콘 업데이트 루프 (더 긴 간격)
func (app *TunnelApp) updateIconLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-app.quitCh:
			return
		case <-ticker.C:
			app.updateTrayIcon()
		}
	}
}

// updateStatus 상태 업데이트
func (app *TunnelApp) updateStatus() {
	if app.manager == nil {
		return
	}

	healthyCount := app.manager.GetHealthyCount()
	totalCount := app.manager.GetTotalCount()

	// 툴팁 업데이트 (안전하게 처리)
	tooltip := fmt.Sprintf("%s - %d/%d connected", version.AppName, healthyCount, totalCount)
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("툴팁 업데이트 에러 (무시됨): %v", r)
			}
		}()
		systray.SetTooltip(tooltip)
	}()

	// 터널 상태 메뉴 아이템들 업데이트
	app.updateStatusItems()
}

// createStatusItems 터널 상태 메뉴 아이템들 생성
func (app *TunnelApp) createStatusItems() {
	if app.manager == nil {
		return
	}

	tunnelStatuses := app.manager.GetTunnelStatuses()
	for _, tunnelStatus := range tunnelStatuses {
		statusText := app.formatTunnelStatus(tunnelStatus)
		item := systray.AddMenuItem(statusText, fmt.Sprintf("Tunnel: %s", tunnelStatus.Name))
		// 상태에 따른 아이콘 설정
		app.setStatusIcon(item, tunnelStatus.Status)
		app.statusItems[tunnelStatus.Name] = item
	}
}

// updateStatusItems 터널 상태 메뉴 아이템들 업데이트
func (app *TunnelApp) updateStatusItems() {
	if app.manager == nil {
		return
	}

	tunnelStatuses := app.manager.GetTunnelStatuses()
	for _, tunnelStatus := range tunnelStatuses {
		if item, exists := app.statusItems[tunnelStatus.Name]; exists {
			statusText := app.formatTunnelStatus(tunnelStatus)
			item.SetTitle(statusText)
			// 상태에 따른 아이콘 업데이트
			app.setStatusIcon(item, tunnelStatus.Status)
		}
	}
}

// updateMenuForConfigReload 설정 리로드 시 메뉴 업데이트
func (app *TunnelApp) updateMenuForConfigReload() {
	if app.manager == nil {
		return
	}

	// 현재 설정의 터널 목록 가져오기
	currentTunnels := make(map[string]bool)
	for _, tunnelConfig := range app.manager.GetConfig().GetEnabledTunnels() {
		currentTunnels[tunnelConfig.Name] = true
	}

	// 기존 상태 항목들 중 제거된 터널들 숨기기
	for name, item := range app.statusItems {
		if !currentTunnels[name] {
			item.Hide()
			delete(app.statusItems, name)
		}
	}

	// 새로 추가된 터널들에 대한 상태 항목 생성
	for name := range currentTunnels {
		if _, exists := app.statusItems[name]; !exists {
			// 새 터널에 대한 상태 항목 생성
			tunnelStatuses := app.manager.GetTunnelStatuses()
			for _, tunnelStatus := range tunnelStatuses {
				if tunnelStatus.Name == name {
					statusText := app.formatTunnelStatus(tunnelStatus)
					item := systray.AddMenuItem(statusText, fmt.Sprintf("Tunnel: %s", name))
					app.setStatusIcon(item, tunnelStatus.Status)
					app.statusItems[name] = item
					break
				}
			}
		}
	}

	// 기존 상태 항목들 업데이트
	app.updateStatusItems()
}

// formatTunnelStatus 터널 상태 포맷팅
func (app *TunnelApp) formatTunnelStatus(status manager.TunnelStatus) string {
	config := status.Config
	var statusText string

	switch status.Status {
	case tunnel.StatusConnected:
		statusText = fmt.Sprintf("● %s (%d) [CONNECTED]",
			status.Name, config.LocalPort)
	case tunnel.StatusConnecting:
		statusText = fmt.Sprintf("⊙ %s (%d) [CONNECTING...]",
			status.Name, config.LocalPort)
	case tunnel.StatusError:
		statusText = fmt.Sprintf("⊗ %s (%d) [ERROR]",
			status.Name, config.LocalPort)
	default:
		statusText = fmt.Sprintf("○ %s (%d) [DISCONNECTED]",
			status.Name, config.LocalPort)
	}

	return statusText
}

// setStatusIcon 메뉴 아이템에 상태에 따른 아이콘 설정 (텍스트 기반)
func (app *TunnelApp) setStatusIcon(item *systray.MenuItem, status tunnel.Status) {
	// Windows systray 라이브러리의 아이콘 설정 문제로 인해
	// 아이콘 설정을 건너뛰고 텍스트만으로 상태 표시
	// 실제 색상 표시는 formatTunnelStatus에서 이모지로 처리됨
}

// showError 에러 메시지 표시 (Windows에서는 메시지 박스 사용)
func (app *TunnelApp) showError(title, message string) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show('%s', '%s')",
				message, title))
		cmd.Run()
	} else {
		log.Printf("ERROR [%s]: %s", title, message)
	}
}

// setInitialTrayIcon 초기 트레이 아이콘 설정 (disconnected)
func (app *TunnelApp) setInitialTrayIcon() {
	// 초기에는 disconnected 아이콘으로 설정
	iconData, err := app.iconAssets.ReadFile("assets/icons/disconnected.ico")
	if err != nil {
		log.Printf("초기 아이콘 로드 실패: %v", err)
		return
	}

	// 아이콘 설정 시도 (에러 무시)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("초기 아이콘 설정 오류 (무시됨): %v", r)
		}
	}()

	systray.SetIcon(iconData)
	app.iconPath = "disconnected"
	log.Printf("초기 트레이 아이콘 설정: disconnected")
}

// updateTrayIcon 상태에 따른 트레이 아이콘 업데이트
func (app *TunnelApp) updateTrayIcon() {
	if app.manager == nil {
		return
	}

	healthyCount := app.manager.GetHealthyCount()
	totalCount := app.manager.GetTotalCount()

	// 상태가 변경되지 않았으면 아이콘 업데이트 스킵
	var newIconName string
	if totalCount == 0 {
		newIconName = "disconnected"
	} else if healthyCount == totalCount {
		newIconName = "connected"
	} else if healthyCount > 0 {
		newIconName = "partial"
	} else {
		newIconName = "disconnected"
	}

	// 현재 아이콘과 같으면 업데이트 스킵 (로그 스팸 방지)
	if app.iconPath == newIconName {
		return
	}

	// ICO 파일만 사용 (Windows 11 호환성) - embed에서 로드
	iconData, err := app.iconAssets.ReadFile("assets/icons/" + newIconName + ".ico")
	if err != nil {
		log.Printf("ICO 아이콘 로드 실패 %s: %v", newIconName, err)
		return
	}

	// Windows 11 호환성을 위한 재시도 로직
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 에러 무시하고 재시도
				}
			}()

			systray.SetIcon(iconData)
			app.iconPath = newIconName
			// 첫 번째 시도에서만 로그 출력 (상태 변경 시에만)
			if i == 0 && (healthyCount == 0 || healthyCount == totalCount) {
				log.Printf("트레이 아이콘 변경: %s (%d/%d)", newIconName, healthyCount, totalCount)
			}
		}()

		// 짧은 대기 후 다음 시도
		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (app *TunnelApp) logTrayStatus() {
	log.Printf("=== %s 애플리케이션 시작됨 ===", version.AppFullName)
}

// CleanupSSHProcesses 남은 SSH 프로세스 정리 (Public)
func (app *TunnelApp) CleanupSSHProcesses() {
	// 여러 방법으로 SSH 프로세스 정리 시도
	cleanupMethods := []struct {
		name string
		cmd  *exec.Cmd
	}{
		{"taskkill /f /im ssh.exe", exec.Command("taskkill", "/f", "/im", "ssh.exe")},
		{"taskkill /f /im OpenSSH", exec.Command("taskkill", "/f", "/im", "OpenSSH")},
		{"wmic process where name='ssh.exe' delete", exec.Command("wmic", "process", "where", "name='ssh.exe'", "delete")},
	}

	for _, method := range cleanupMethods {
		// 조용히 실행 (출력 숨김)
		method.cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		}

		if err := method.cmd.Run(); err != nil {
			log.Printf("SSH 프로세스 정리 방법 '%s' 실패: %v", method.name, err)
		} else {
			log.Printf("SSH 프로세스 정리 방법 '%s' 성공", method.name)
		}
	}

	log.Println("SSH 프로세스 정리 완료")
}
