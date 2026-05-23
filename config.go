package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Port       string `json:"port"`
	Env        string `json:"env"`
	AdminPwd   string `json:"admin_pwd"`
	DB         string `json:"db"`
	COSBucket  string `json:"cos_bucket"`
	COSRegion  string `json:"cos_region"`
	COSSecretID string `json:"cos_secret_id"`
	COSSecretKey string `json:"cos_secret_key"`
	// FaceChain train-free template URLs (male/female)
	FaceChainTemplateMale   string `json:"facechain_template_male"`
	FaceChainTemplateFemale string `json:"facechain_template_female"`
	// HivisionIDPhotos endpoint
	HivisionURL string `json:"hivision_url"`
	// Frontend serve path (prod only)
	FrontendDir string `json:"frontend_dir,omitempty"`
}

var cfg Config

func loadConfig() error {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	// Try env-specific config first, fall back to default
	candidates := []string{
		filepath.Join("config", fmt.Sprintf("config.%s.json", env)),
		filepath.Join("config", "config.json"),
		fmt.Sprintf("config.%s.json", env),
		"config.json",
	}

	var data []byte
	var err error
	for _, path := range candidates {
		data, err = os.ReadFile(path)
		if err == nil {
			logger.Printf("加载配置: %s", path)
			break
		}
	}
	if err != nil {
		return fmt.Errorf("未找到配置文件，请设置 config.json 或 config/config.%s.json", env)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	return nil
}
