package config

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
	UploadDir string
	Port      = ":8080"
	Cfg       Config
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

// LoadConfig 加载配置文件
func LoadConfig() {
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

		Cfg = defaultConfig
		log.Printf("已创建默认配置文件: %s", configFile)
		return
	}

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	if err := json.Unmarshal(data, &Cfg); err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}

	// 如果配置文件中存储目录为空，使用默认下载目录
	if Cfg.StorageDir == "" {
		Cfg.StorageDir = getDefaultDownloadDir()
		// 保存更新后的配置
		SaveConfig()
	}

	// 如果配置文件中端口为空，使用默认端口
	if Cfg.Port == "" {
		Cfg.Port = ":8080"
	}

	// 如果配置文件中根路由为空，使用默认根路由
	if Cfg.RootPath == "" {
		Cfg.RootPath = "/"
	}

	UploadDir = Cfg.StorageDir
	Port = Cfg.Port

	// 确保目录存在
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		log.Fatalf("无法创建存储目录 %s: %v", UploadDir, err)
	}

	absPath, _ := filepath.Abs(UploadDir)
	log.Printf("存储目录已设置为: %s", absPath)
	log.Printf("服务器端口: %s", Port)
	log.Printf("根路由已设置为: %s", Cfg.RootPath)
}

// SaveConfig 保存配置文件
func SaveConfig() {
	data, err := json.MarshalIndent(Cfg, "", "  ")
	if err != nil {
		log.Printf("无法序列化配置: %v", err)
		return
	}

	if err := os.WriteFile("config.json", data, 0644); err != nil {
		log.Printf("无法保存配置文件: %v", err)
	}
}
