package sync

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shinychan95/Chan/notion"
	"github.com/shinychan95/Chan/utils"
)

type SyncResult struct {
	Success    bool
	Message    string
	Error      error
	PostCount  int
	ImageCount int
	Duration   time.Duration
	Timestamp  time.Time
}

type SyncStatus struct {
	IsRunning  bool
	Progress   string
	LastResult *SyncResult
}

type BlogSyncer struct {
	config *utils.Config
	status *SyncStatus
	mutex  sync.RWMutex
}

func NewBlogSyncer(configPath string) (*BlogSyncer, error) {
	config, err := utils.ReadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &BlogSyncer{
		config: config,
		status: &SyncStatus{
			IsRunning: false,
		},
	}, nil
}

// NewBlogSyncerWithConfig creates a BlogSyncer with a config object
// Expects a config object with ToUtilsConfig() method
func NewBlogSyncerWithConfig(cfg interface{}) (*BlogSyncer, error) {
	var config *utils.Config

	// Check if config has ToUtilsConfig method (config.Config type)
	if c, ok := cfg.(interface{ ToUtilsConfig() *utils.Config }); ok {
		config = c.ToUtilsConfig()
	} else if c, ok := cfg.(*utils.Config); ok {
		config = c
	} else {
		return nil, fmt.Errorf("unsupported config type")
	}

	return &BlogSyncer{
		config: config,
		status: &SyncStatus{
			IsRunning: false,
		},
	}, nil
}

func (bs *BlogSyncer) GetStatus() SyncStatus {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	return *bs.status
}

func (bs *BlogSyncer) updateStatus(isRunning bool, progress string) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.status.IsRunning = isRunning
	bs.status.Progress = progress
	log.Printf("🦤 Sync status: %s", progress)
}

func (bs *BlogSyncer) setResult(result *SyncResult) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.status.LastResult = result
	bs.status.IsRunning = false
	bs.status.Progress = ""
}

func (bs *BlogSyncer) SyncToBlog() *SyncResult {
	startTime := time.Now()

	// 이미 실행 중인지 확인
	if bs.GetStatus().IsRunning {
		return &SyncResult{
			Success:   false,
			Message:   "동기화가 이미 진행 중입니다",
			Error:     fmt.Errorf("sync already in progress"),
			Timestamp: time.Now(),
		}
	}

	bs.updateStatus(true, "초기화 중...")

	// UUID 검증
	rootID, err := utils.CheckUUIDv4Format(bs.config.RootID)
	if err != nil {
		result := &SyncResult{
			Success:   false,
			Message:   "잘못된 Root ID 형식",
			Error:     err,
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
		bs.setResult(result)
		return result
	}

	// Notion 초기화
	bs.updateStatus(true, "Notion 연결 중...")
	notion.Init(bs.config.ApiKey, bs.config.PostDir, bs.config.ImgDir, bs.config.DBPath)
	defer notion.Close()

	// 기존 포스트 삭제
	bs.updateStatus(true, "기존 포스트 정리 중...")
	err = bs.clearExistingPosts()
	if err != nil {
		result := &SyncResult{
			Success:   false,
			Message:   "기존 포스트 삭제 실패",
			Error:     err,
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
		bs.setResult(result)
		return result
	}

	// 동기화 실행
	bs.updateStatus(true, "Notion에서 데이터 가져오는 중...")
	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	notion.HandleCollectionView(rootID, &wg, errCh)

	// 이미지 다운로드 대기
	bs.updateStatus(true, "이미지 다운로드 중...")
	go func() {
		wg.Wait()
		close(errCh)
	}()

	var syncError error
	for err := range errCh {
		if err != nil {
			syncError = err
			break
		}
	}

	if syncError != nil {
		result := &SyncResult{
			Success:   false,
			Message:   "동기화 중 오류 발생",
			Error:     syncError,
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
		bs.setResult(result)
		return result
	}

	// Git 커밋 및 푸시
	bs.updateStatus(true, "블로그에 배포 중...")
	err = bs.gitCommitAndPush()
	if err != nil {
		result := &SyncResult{
			Success:   false,
			Message:   "Git 배포 실패",
			Error:     err,
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}
		bs.setResult(result)
		return result
	}

	// 성공 결과
	result := &SyncResult{
		Success:   true,
		Message:   "블로그 동기화 완료",
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
	}
	bs.setResult(result)
	return result
}

func (bs *BlogSyncer) clearExistingPosts() error {
	postDir := bs.config.PostDir
	files, err := filepath.Glob(filepath.Join(postDir, "*.md"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			log.Printf("Warning: could not remove file %s: %v", file, err)
		}
	}

	return nil
}

func (bs *BlogSyncer) gitCommitAndPush() error {
	// blog 저장소 경로 추출 (post_directory의 상위 디렉토리)
	repoPath := filepath.Dir(bs.config.PostDir)

	// Git 설정을 조정하여 큰 파일 업로드 문제 해결
	// HTTP 버퍼 크기 증가 (기본값: 1MB -> 500MB)
	cmd := exec.Command("git", "config", "http.postBuffer", "524288000")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: git config http.postBuffer 실패: %v", err)
	}

	// 압축 레벨 조정 (기본값: 6 -> 0, 압축 비활성화로 속도 향상)
	cmd = exec.Command("git", "config", "core.compression", "0")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: git config core.compression 실패: %v", err)
	}

	// HTTP 타임아웃 증가 (기본값: 60초 -> 300초)
	cmd = exec.Command("git", "config", "http.lowSpeedLimit", "0")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: git config http.lowSpeedLimit 실패: %v", err)
	}

	cmd = exec.Command("git", "config", "http.lowSpeedTime", "300")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: git config http.lowSpeedTime 실패: %v", err)
	}

	// git add .
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add 실패: %w", err)
	}

	// git commit
	commitMsg := fmt.Sprintf("Update blog content - %s", time.Now().Format("2006-01-02 15:04:05"))
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		// 변경사항이 없는 경우는 에러로 처리하지 않음
		if !strings.Contains(err.Error(), "nothing to commit") {
			return fmt.Errorf("git commit 실패: %w", err)
		}
		// 커밋할 내용이 없으면 푸시도 필요 없으므로 성공으로 간주
		return nil
	}

	// GitHub 토큰이 설정되어 있으면 원격 URL을 업데이트
	if bs.config.GitHubToken != "" && bs.config.GitHubRepo != "" {
		// 표준 형식으로 원격 URL 설정 (문서 권장 방식)
		remoteURL := fmt.Sprintf("https://%s@github.com/%s.git", bs.config.GitHubToken, bs.config.GitHubRepo)
		cmd = exec.Command("git", "remote", "set-url", "origin", remoteURL)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git remote set-url 실패: %w", err)
		}
	}

	// git push (큰 파일 처리를 위한 추가 옵션 포함)
	cmd = exec.Command("git", "push", "--verbose")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push 실패: %w\n%s", err, string(output))
	}

	return nil
}
