package main

import (
	"embed"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"tunnels/internal/app"

	"github.com/getlantern/systray"
)

//go:embed assets/icons/*.ico
var iconAssets embed.FS

func main() {
	// Windows에서 콘솔 창 숨기기
	if runtime.GOOS == "windows" {
		hideConsoleWindow()
	}

	// 콘솔 출력 완전 차단 (가장 먼저)
	log.SetOutput(io.Discard)
	os.Stdout = nil
	os.Stderr = nil

	// 로그 파일 크기 확인 및 로테이션 (조용히)
	logPath := "tunnels.log"
	rotateLogIfNeeded(logPath)

	// 로그 파일로 출력 설정
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer logFile.Close()
		// 로그를 파일로만 출력하도록 설정
		log.SetOutput(logFile)
	}

	// 설정 파일 경로
	configPath := "tunnels.conf"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 절대 경로로 변환
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		// 에러 발생 시에도 조용히 종료
		return
	}

	// 앱 인스턴스 생성
	app := app.NewTunnelApp(absConfigPath, iconAssets)

	// 시그널 처리 (Ctrl+C 등)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("시그널 수신 - 애플리케이션 종료 중...")
		// SSH 프로세스 정리
		app.CleanupSSHProcesses()
		os.Exit(0)
	}()

	// 시스템 트레이 시작
	systray.Run(app.OnReady, app.OnExit)
}

// rotateLogIfNeeded 로그 파일 크기가 1MB를 초과하면 로테이션
func rotateLogIfNeeded(logPath string) error {
	// 로그 파일이 존재하지 않으면 로테이션 불필요
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return nil
	}

	// 파일 크기 확인
	fileInfo, err := os.Stat(logPath)
	if err != nil {
		return err
	}

	// 1MB (1024*1024 bytes) 이상이면 로테이션
	if fileInfo.Size() > 1024*1024 {
		// 백업 파일 생성
		backupPath := logPath + ".old"
		if err := os.Rename(logPath, backupPath); err != nil {
			return err
		}
	}

	return nil
}

// hideConsoleWindow Windows에서 콘솔 창 숨기기
func hideConsoleWindow() {
	if runtime.GOOS != "windows" {
		return
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")

	procGetConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	procShowWindow := user32.NewProc("ShowWindow")
	procFreeConsole := kernel32.NewProc("FreeConsole")

	// 콘솔 창 숨기기
	consoleWindow, _, _ := procGetConsoleWindow.Call()
	if consoleWindow != 0 {
		procShowWindow.Call(consoleWindow, 0) // SW_HIDE = 0
	}

	// 콘솔 해제 (더 강력한 방법)
	procFreeConsole.Call()
}
