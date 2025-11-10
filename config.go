package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	StorageDir string `json:"storage_dir"`
	Port       string `json:"port"`
	RootPath   string `json:"root_path"`
}

var (
	uploadDir string
	port      = ":8080"
	config    Config
)

// 获取默认下载目录
func getDefaultDownloadDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("无法获取用户主目录，使用当前目录: %v", err)
		return "./downloads"
	}

	var downloadDir string
	switch runtime.GOOS {
	case "windows":
		downloadDir = filepath.Join(homeDir, "Downloads")
	case "darwin": // macOS
		downloadDir = filepath.Join(homeDir, "Downloads")
	case "linux":
		downloadDir = filepath.Join(homeDir, "Downloads")
	default:
		downloadDir = filepath.Join(homeDir, "Downloads")
	}

	return downloadDir
}

// 加载配置文件
func loadConfig() {
	configFile := "config.json"

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := Config{
			StorageDir: getDefaultDownloadDir(),
			Port:       ":8080",
			RootPath:   "/",
		}

		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			log.Fatalf("无法创建默认配置: %v", err)
		}

		if err := os.WriteFile(configFile, data, 0644); err != nil {
			log.Fatalf("无法写入配置文件: %v", err)
		}

		config = defaultConfig
		log.Printf("已创建默认配置文件: %s", configFile)
		return
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}

	// 如果配置文件中存储目录为空，使用默认下载目录
	if config.StorageDir == "" {
		config.StorageDir = getDefaultDownloadDir()
		// 保存更新后的配置
		saveConfig()
	}

	// 如果配置文件中端口为空，使用默认端口
	if config.Port == "" {
		config.Port = ":8080"
	}

	// 如果配置文件中根路由为空，使用默认根路由
	if config.RootPath == "" {
		config.RootPath = "/"
	}

	uploadDir = config.StorageDir
	port = config.Port

	// 确保目录存在
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("无法创建存储目录 %s: %v", uploadDir, err)
	}

	absPath, _ := filepath.Abs(uploadDir)
	log.Printf("存储目录已设置为: %s", absPath)
	log.Printf("服务器端口: %s", port)
	log.Printf("根路由已设置为: %s", config.RootPath)
}

// 保存配置文件
func saveConfig() {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("无法序列化配置: %v", err)
		return
	}

	if err := os.WriteFile("config.json", data, 0644); err != nil {
		log.Printf("无法保存配置文件: %v", err)
	}
}
