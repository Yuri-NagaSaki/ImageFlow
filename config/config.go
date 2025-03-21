package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 存储应用配置
type Config struct {
	ServerAddr    string `json:"server_addr"`
	ImageBasePath string `json:"image_base_path"`
	AvifSupport   bool   `json:"avif_support"`
	APIKey        string // 添加 API Key
	MaxUploadCount int    `json:"max_upload_count"` // 最大允许一次性上传图片数量
	ImageQuality  int    `json:"image_quality"`   // 图片转换质量(1-100)
	WorkerThreads int    `json:"worker_threads"` // 并行工作线程数
}

// Load 从配置文件加载配置
func Load() (*Config, error) {
	// 默认配置
	cfg := &Config{
		ServerAddr:    "0.0.0.0:8686",
		ImageBasePath: os.Getenv("LOCAL_STORAGE_PATH"),
		AvifSupport:   true,
		MaxUploadCount: 10,  // 默认一次最多上传 10 张图片
		ImageQuality:  80,  // 默认质量 80
		WorkerThreads: 4,   // 默认 4 个线程
	}

	// 如果没有设置 LOCAL_STORAGE_PATH，使用默认值
	if cfg.ImageBasePath == "" {
		cfg.ImageBasePath = "static/images"
	}

	// 确保路径是相对路径
	if !filepath.IsAbs(cfg.ImageBasePath) {
		cfg.ImageBasePath = filepath.Join(".", cfg.ImageBasePath)
	}

	// 尝试加载 .env 文件，但不强制要求
	_ = godotenv.Load()

	// 从环境变量获取配置
	cfg.APIKey = os.Getenv("API_KEY")

	// 从环境变量获取最大一次性上传图片数量
	if maxCount := os.Getenv("MAX_UPLOAD_COUNT"); maxCount != "" {
		if count, err := strconv.Atoi(maxCount); err == nil && count > 0 {
			cfg.MaxUploadCount = count
		}
	}

	// 从环境变量获取图片质量
	if quality := os.Getenv("IMAGE_QUALITY"); quality != "" {
		if q, err := strconv.Atoi(quality); err == nil && q > 0 && q <= 100 {
			cfg.ImageQuality = q
		}
	}

	// 从环境变量获取工作线程数
	if workers := os.Getenv("WORKER_THREADS"); workers != "" {
		if w, err := strconv.Atoi(workers); err == nil && w > 0 {
			cfg.WorkerThreads = w
		}
	}

	// 如果配置文件存在，从文件加载其他配置
	if _, err := os.Stat("config/config.json"); err == nil {
		file, err := os.Open("config/config.json")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
