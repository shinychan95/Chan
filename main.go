package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void forceAccessoryModeEarly() {
    // NSApplication이 아직 생성되지 않았다면 생성하고 즉시 Accessory 모드로 설정
    NSApplication *app = [NSApplication sharedApplication];
    [app setActivationPolicy:NSApplicationActivationPolicyAccessory];
}
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3" // SQLite 드라이버 로드
	"github.com/shinychan95/Chan/config"
	blogsync "github.com/shinychan95/Chan/sync"
)

var (
	syncer           *blogsync.BlogSyncer
	cfg              *config.Config
	isConfigured     bool
	fyneApp          fyne.App
	mainWindow       fyne.Window
	windowMutex      sync.Mutex
	autoStartEnabled bool
	forceQuit        bool // 강제 종료 플래그
)

func main() {
	// macOS에서 NSApplication 생성과 동시에 즉시 Accessory 모드 적용
	switch runtime.GOOS {
	case "darwin":
		C.forceAccessoryModeEarly() // 핵심 해결책
	}

	fyneApp = app.NewWithID("com.shinychan95.Chan")
	fyneApp.SetIcon(resourceDodoPng)
	checkAutoStartStatus()

	// 앱 생명주기 관리
	fyneApp.Lifecycle().SetOnStopped(func() {
		if !forceQuit {
			log.Println("앱 종료 시도가 감지되었습니다. 트레이에서만 종료 가능합니다.")
			showNotification("알림", "앱을 종료하려면 메뉴바에서 'Quit'을 선택해주세요.")
			return
		}
		log.Println("앱이 정상적으로 종료됩니다.")
	})

	// 트레이 아이콘 설정
	if desk, ok := fyneApp.(desktop.App); ok {
		desk.SetSystemTrayIcon(resourceDodoPng)
	}

	// 설정 로드 및 트레이 메뉴 설정
	loadConfiguration()
	refreshTrayMenu()

	// 앱 실행
	fyneApp.Run()
}

func refreshTrayMenu() {
	if desk, ok := fyneApp.(desktop.App); ok {
		menuItems := []*fyne.MenuItem{
			fyne.NewMenuItem("Chan 열기", func() {
				showMainWindow()
			}),
		}

		if isConfigured {
			menuItems = append(menuItems, fyne.NewMenuItemSeparator(), fyne.NewMenuItem("Notion 동기화", func() {
				if isConfigured {
					go handleSync()
					showNotification("시작", "동기화를 시작합니다...")
				} else {
					showNotification("설정 필요", "먼저 설정을 완료해주세요")
				}
			}))
		}

		menuItems = append(menuItems, fyne.NewMenuItemSeparator(), fyne.NewMenuItem("Quit", func() {
			forceQuit = true // 강제 종료 플래그 설정
			fyneApp.Quit()
		}))

		trayMenu := fyne.NewMenu("Chan", menuItems...)
		desk.SetSystemTrayMenu(trayMenu)
		desk.SetSystemTrayIcon(resourceDodoPng)
	}
}

func showMainWindow() {
	windowMutex.Lock()
	defer windowMutex.Unlock()

	if mainWindow == nil {
		mainWindow = fyneApp.NewWindow("Chan")
		mainWindow.Resize(fyne.NewSize(600, 500))
		mainWindow.CenterOnScreen()
		mainWindow.SetContent(createSettingsUI(mainWindow))
		mainWindow.SetCloseIntercept(func() {
			mainWindow.Hide()
		})

		// ESC 키로 창 닫기
		mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
			if ke.Name == fyne.KeyEscape {
				mainWindow.Hide()
			}
		})
	}
	mainWindow.Show()
	mainWindow.RequestFocus()
}

func loadConfiguration() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Printf("설정 로드 실패: %v", err)
		cfg = config.CreateDefaultConfig()
		if err := config.SaveConfig(cfg); err != nil {
			log.Printf("기본 설정 저장 실패: %v", err)
		}
		isConfigured = false
		showNotification("알림", "🦤 첫 실행입니다. 설정을 완료해주세요!")
		return
	}
	errors := config.ValidateConfig(cfg)
	if len(errors) > 0 {
		log.Printf("설정 오류: %v", errors)
		isConfigured = false
		showNotification("설정 필요", strings.Join(errors[:1], ""))
		return
	}
	syncer, err = blogsync.NewBlogSyncerWithConfig(cfg)
	if err != nil {
		log.Printf("동기화 객체 생성 실패: %v", err)
		isConfigured = false
		showNotification("오류", "동기화 초기화에 실패했습니다")
		return
	}
	isConfigured = true
	showNotification("준비 완료", "🦤 Notion Blog가 준비되었습니다!")
}

func createSettingsUI(win fyne.Window) fyne.CanvasObject {
	// General settings tab
	autoStartCheck := widget.NewCheck("로그인 시 자동 실행", func(checked bool) {
		if checked {
			enableAutoStart()
		} else {
			disableAutoStart()
		}
	})
	autoStartCheck.SetChecked(autoStartEnabled)
	generalTab := container.NewVBox(
		widget.NewLabelWithStyle("일반 설정", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		autoStartCheck,
	)

	// Notion settings tab
	notionTab := createNotionSettingsUI(win)

	tabs := container.NewAppTabs(
		container.NewTabItem("일반", generalTab),
		container.NewTabItem("Notion", notionTab),
	)

	return tabs
}

func createNotionSettingsUI(win fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("🦤 Notion 설정")
	title.TextStyle.Bold = true
	var statusText string
	if isConfigured {
		statusText = "✅ 설정 완료"
	} else {
		statusText = "⚠️ 설정이 필요합니다"
	}
	status := widget.NewLabel(statusText)

	// Notion API Key
	apiKeyEntry := widget.NewPasswordEntry()
	apiKeyEntry.SetPlaceHolder("Notion API 키를 입력하세요")
	if cfg != nil && cfg.ApiKey != "" {
		apiKeyEntry.SetText(cfg.ApiKey)
	}

	// Notion DB Path
	dbPathEntry := widget.NewEntry()
	dbPathEntry.SetPlaceHolder("/Users/user/Library/Application Support/Notion/notion.db")
	if cfg != nil && cfg.DBPath != "" {
		dbPathEntry.SetText(cfg.DBPath)
	} else {
		// 기본값 설정 (config 패키지의 기본 설정 사용)
		defaultCfg := config.CreateDefaultConfig()
		if defaultCfg.DBPath != "" {
			dbPathEntry.SetText(defaultCfg.DBPath)
		}
	}
	selectDBButton := widget.NewButton("DB 파일 선택", func() {
		if runtime.GOOS == "darwin" {
			script := `POSIX path of (choose file with prompt "Notion DB 파일 선택" of type {"db"})`
			out, err := exec.Command("osascript", "-e", script).Output()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
					return // 사용자가 취소함
				}
				showNotification("오류", fmt.Sprintf("파일 선택 중 오류 발생: %v", err))
				return
			}
			dbPathEntry.SetText(strings.TrimSpace(string(out)))
		} else {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					showNotification("오류", "파일 선택 중 오류가 발생했습니다")
					return
				}
				if reader == nil {
					return // 사용자가 취소함
				}
				defer reader.Close()

				selectedPath := reader.URI().Path()
				if !strings.HasSuffix(selectedPath, ".db") {
					showNotification("경고", "DB 파일(.db)을 선택해주세요")
					return
				}
				dbPathEntry.SetText(selectedPath)
			}, win)
		}
	})

	// Collection View ID
	dbIDEntry := widget.NewEntry()
	dbIDEntry.SetPlaceHolder("Collection View ID를 입력하세요 (URL의 마지막 부분)")
	if cfg != nil && cfg.RootID != "" {
		dbIDEntry.SetText(cfg.RootID)
	}

	// Output Directory
	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetPlaceHolder("출력 디렉토리 경로 (Post Directory)")
	if cfg != nil && cfg.PostDir != "" {
		outputDirEntry.SetText(cfg.PostDir)
	}
	selectDirButton := widget.NewButton("폴더 선택", func() {
		if runtime.GOOS == "darwin" {
			script := `POSIX path of (choose folder with prompt "출력 디렉토리 선택")`
			out, err := exec.Command("osascript", "-e", script).Output()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
					return // 사용자가 취소함
				}
				showNotification("오류", fmt.Sprintf("폴더 선택 중 오류 발생: %v", err))
				return
			}
			outputDirEntry.SetText(strings.TrimSpace(string(out)))
		} else {
			dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
				if err != nil {
					showNotification("오류", "폴더 선택 중 오류가 발생했습니다")
					return
				}
				if list == nil {
					return // 사용자가 취소함
				}
				selectedPath := list.Path()
				outputDirEntry.SetText(selectedPath)
			}, win)
		}
	})

	// GitHub Token
	githubTokenEntry := widget.NewPasswordEntry()
	githubTokenEntry.SetPlaceHolder("GitHub Personal Access Token")
	if cfg != nil && cfg.GitHubToken != "" {
		githubTokenEntry.SetText(cfg.GitHubToken)
	}

	// GitHub Repo URL
	githubRepoEntry := widget.NewEntry()
	githubRepoEntry.SetPlaceHolder("사용자이름/저장소이름 (예: shinychan95/my-blog)")
	if cfg != nil && cfg.GitHubRepo != "" {
		githubRepoEntry.SetText(cfg.GitHubRepo)
	}

	var syncButton *widget.Button
	saveButton := widget.NewButton("설정 저장", func() {
		if cfg == nil {
			cfg = config.CreateDefaultConfig()
		}
		cfg.ApiKey = apiKeyEntry.Text
		cfg.DBPath = dbPathEntry.Text
		cfg.RootID = dbIDEntry.Text
		cfg.PostDir = outputDirEntry.Text
		cfg.GitHubToken = githubTokenEntry.Text
		cfg.GitHubRepo = githubRepoEntry.Text
		if err := config.SaveConfig(cfg); err != nil {
			showNotification("오류", fmt.Sprintf("설정 저장 실패: %v", err))
			return
		}
		loadConfiguration()
		refreshTrayMenu()
		if isConfigured {
			status.SetText("✅ 설정 완료")
			if syncButton != nil {
				syncButton.Enable()
			}
			showNotification("완료", "설정이 저장되었습니다!")
		} else {
			status.SetText("⚠️ 설정 확인이 필요합니다")
			showNotification("경고", "설정을 확인해주세요")
		}
	})

	syncButton = widget.NewButton("🦤 동기화 실행", func() {
		if isConfigured {
			go handleSync()
			showNotification("시작", "동기화를 시작합니다...")
		} else {
			showNotification("설정 필요", "먼저 설정을 완료해주세요")
		}
	})
	if !isConfigured {
		syncButton.Disable()
	}

	form := container.NewVBox(
		title,
		status,
		widget.NewSeparator(),
		widget.NewLabel("Notion API 키:"),
		apiKeyEntry,
		widget.NewLabel("Notion DB 경로:"),
		container.NewBorder(nil, nil, nil, selectDBButton, dbPathEntry),
		widget.NewLabel("Collection View ID:"),
		dbIDEntry,
		widget.NewLabel("출력 디렉토리:"),
		container.NewBorder(nil, nil, nil, selectDirButton, outputDirEntry),
		widget.NewSeparator(),
		widget.NewLabel("GitHub Token:"),
		githubTokenEntry,
		widget.NewLabel("GitHub Repository:"),
		githubRepoEntry,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, saveButton, syncButton),
	)
	return form
}

func handleSync() {
	if !isConfigured || syncer == nil {
		showNotification("오류", "설정을 먼저 완료해주세요")
		return
	}
	status := syncer.GetStatus()
	if status.IsRunning {
		showNotification("알림", "이미 동기화가 진행 중입니다")
		return
	}
	go func() {
		// Panic 복구
		defer func() {
			if r := recover(); r != nil {
				log.Printf("동기화 중 패닉 발생: %v", r)
				// 스택 트레이스도 로깅할 수 있습니다.
				debug.PrintStack()
				showNotification("치명적 오류", "동기화 중 심각한 오류가 발생했습니다. 로그를 확인해주세요.")
			}
		}()

		showNotification("시작", "🦤 블로그 동기화를 시작합니다")
		result := syncer.SyncToBlog()
		if result.Success {
			msg := fmt.Sprintf("동기화 완료! (소요시간: %s)", formatDuration(result.Duration))
			showNotification("완료", msg)
		} else {
			showNotification("오류", result.Message)
			log.Printf("동기화 실패: %v", result.Error)
		}
	}()
}

func checkAutoStartStatus() {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Launch Agent 확인
		plistPath := getLaunchAgentPath()
		if _, err := os.Stat(plistPath); err == nil {
			autoStartEnabled = true
		} else {
			autoStartEnabled = false
		}
	case "windows":
		// Windows: 레지스트리 또는 시작프로그램 폴더 확인 (구현 필요)
		autoStartEnabled = false // 기본값
	}
}

func enableAutoStart() {
	switch runtime.GOOS {
	case "darwin":
		plistPath := getLaunchAgentPath()
		ex, err := os.Executable()
		if err != nil {
			log.Printf("실행 경로 얻기 실패: %v", err)
			return
		}
		plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.shinychan95.Chan</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, ex)
		err = os.WriteFile(plistPath, []byte(plistContent), 0644)
		if err != nil {
			log.Printf("Launch Agent 파일 생성 실패: %v", err)
			showNotification("오류", "자동 실행 설정에 실패했습니다.")
			return
		}
		autoStartEnabled = true
		showNotification("성공", "로그인 시 자동 실행이 활성화되었습니다.")
	}
}

func disableAutoStart() {
	switch runtime.GOOS {
	case "darwin":
		plistPath := getLaunchAgentPath()
		err := os.Remove(plistPath)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Launch Agent 파일 삭제 실패: %v", err)
			showNotification("오류", "자동 실행 해제에 실패했습니다.")
			return
		}
		autoStartEnabled = false
		showNotification("성공", "로그인 시 자동 실행이 비활성화되었습니다.")
	}
}

func getLaunchAgentPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("홈 디렉토리 얻기 실패: %v", err)
		return ""
	}
	return filepath.Join(home, "Library", "LaunchAgents", "com.shinychan95.Chan.plist")
}

func showNotification(title, message string) {
	if fyneApp != nil {
		fyneApp.SendNotification(&fyne.Notification{Title: title, Content: message})
	}
}

func formatDuration(d time.Duration) string {
	return d.Round(time.Second).String()
}
