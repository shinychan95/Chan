package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBPath      string `json:"db_path"`
	ApiKey      string `json:"api_key"`
	PostDir     string `json:"post_directory"`
	ImgDir      string `json:"image_directory"`
	RootID      string `json:"root_id"`
	GitHubToken string `json:"github_token"`
	GitHubRepo  string `json:"github_repo"`
}

func ReadConfig(configPath string) (*Config, error) {
	// Check if file exists
	_, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Open the file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the JSON file into a Config struct
	var cfg Config
	dec := json.NewDecoder(file)
	err = dec.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// 만약 notion.db 경로값 없을 경우, 동적으로 파악
	if cfg.DBPath == "" {
		cfg.DBPath = FindNotionDBPath()
	}

	// Check if all fields are present
	if cfg.DBPath == "" || cfg.ApiKey == "" || cfg.PostDir == "" || cfg.ImgDir == "" || cfg.RootID == "" {
		return nil, fmt.Errorf("missing required fields in config file: %s", configPath)
	}

	return &cfg, nil
}
