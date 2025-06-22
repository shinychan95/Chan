package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/shinychan95/Chan/utils"
)

type Config struct {
	DBPath      string `json:"db_path"`
	ApiKey      string `json:"api_key"`
	PostDir     string `json:"post_directory"`
	ImgDir      string `json:"image_directory"`
	RootID      string `json:"root_id"`
	GitHubToken string `json:"github_token"`
	GitHubRepo  string `json:"github_repo"` // 예: "shinychan95/shinychan95.github.io"
}

// GetConfigPath returns the path to the config file in user's Application Support
func GetConfigPath() string {
	var configDir string

	if runtime.GOOS == "darwin" {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, "Library", "Application Support", "Chan")
	} else {
		// 다른 OS의 경우 홈 디렉토리/.chan 사용
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".chan")
	}

	// 디렉토리가 없으면 생성
	os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, "config.json")
}

// LoadConfig loads configuration from the user's Application Support directory
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	// 파일이 존재하지 않으면 기본 설정 반환
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// migrateFromOldPath migrates config from old "Notion Blog" directory to new "Chan" directory
func migrateFromOldPath() {
	if runtime.GOOS != "darwin" {
		return
	}

	homeDir, _ := os.UserHomeDir()
	oldConfigPath := filepath.Join(homeDir, "Library", "Application Support", "Notion Blog", "config.json")
	newConfigPath := GetConfigPath()

	// 새 경로에 이미 설정 파일이 있으면 마이그레이션하지 않음
	if _, err := os.Stat(newConfigPath); err == nil {
		return
	}

	// 기존 설정 파일이 있으면 새 경로로 복사
	if _, err := os.Stat(oldConfigPath); err == nil {
		// 새 디렉토리 생성
		newDir := filepath.Dir(newConfigPath)
		os.MkdirAll(newDir, 0755)

		// 파일 복사
		input, err := os.ReadFile(oldConfigPath)
		if err == nil {
			err = os.WriteFile(newConfigPath, input, 0644)
			if err == nil {
				// 성공적으로 마이그레이션되면 기존 파일 삭제
				os.Remove(oldConfigPath)
				// 빈 디렉토리도 삭제 시도
				os.Remove(filepath.Dir(oldConfigPath))
			}
		}
	}
}

// SaveConfig saves configuration to the user's Application Support directory
func SaveConfig(cfg *Config) error {
	configPath := GetConfigPath()

	// 디렉토리 생성
	dir := filepath.Dir(configPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

// CreateDefaultConfig creates a default configuration with auto-detected values
func CreateDefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		DBPath:  findNotionDBPath(), // macOS 기본 경로 자동 감지
		ApiKey:  "",
		PostDir: filepath.Join(homeDir, "github", "shinychan95.github.io", "_posts"),
		ImgDir:  filepath.Join(homeDir, "github", "shinychan95.github.io", "assets", "pages"),
		RootID:  "",
	}
}

// ValidateConfig checks if the configuration is valid and complete
func ValidateConfig(cfg *Config) []string {
	var errors []string

	if cfg.ApiKey == "" {
		errors = append(errors, "Notion API Key가 설정되지 않았습니다")
	}

	if cfg.RootID == "" {
		errors = append(errors, "Collection View ID가 설정되지 않았습니다")
	}

	if cfg.DBPath == "" {
		errors = append(errors, "Notion DB 경로가 설정되지 않았습니다")
	} else if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("Notion DB가 존재하지 않습니다: %s", cfg.DBPath))
	}

	if cfg.PostDir == "" {
		errors = append(errors, "Post Directory가 설정되지 않았습니다")
	} else if _, err := os.Stat(cfg.PostDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("Post Directory가 존재하지 않습니다: %s", cfg.PostDir))
	}

	if cfg.ImgDir == "" {
		errors = append(errors, "Image Directory가 설정되지 않았습니다")
	}

	return errors
}

// OpenConfigInEditor opens the config file in the default text editor
func OpenConfigInEditor() error {
	configPath := GetConfigPath()

	if runtime.GOOS == "darwin" {
		return exec.Command("open", "-t", configPath).Run()
	}

	// 다른 OS에서는 기본 편집기 사용
	return exec.Command("xdg-open", configPath).Run()
}

// ToUtilsConfig converts config.Config to utils.Config for compatibility
func (c *Config) ToUtilsConfig() *utils.Config {
	return &utils.Config{
		DBPath:      c.DBPath,
		ApiKey:      c.ApiKey,
		PostDir:     c.PostDir,
		ImgDir:      c.ImgDir,
		RootID:      c.RootID,
		GitHubToken: c.GitHubToken,
		GitHubRepo:  c.GitHubRepo,
	}
}

// findNotionDBPath attempts to find the Notion database automatically
func findNotionDBPath() string {
	if runtime.GOOS != "darwin" {
		return ""
	}

	homeDir, _ := os.UserHomeDir()
	possiblePaths := []string{
		filepath.Join(homeDir, "Library", "Application Support", "Notion", "notion.db"),
		filepath.Join(homeDir, "Library", "Application Support", "Notion Desktop", "notion.db"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
