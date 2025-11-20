package models

import "time"

type FileInfo struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	IsDir     bool      `json:"isDir"`
	Extension string    `json:"extension"`
	Path      string    `json:"path,omitempty"` // 相对路径
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Speed   *SpeedInfo  `json:"speed,omitempty"` // 速度信息
}

type SpeedInfo struct {
	AverageSpeed float64 `json:"averageSpeed"` // 平均速度（字节/秒）
	CurrentSpeed float64 `json:"currentSpeed"` // 当前速度（字节/秒）
	TotalBytes   int64   `json:"totalBytes"`   // 总字节数
	Duration     string  `json:"duration"`     // 耗时
	SpeedText    string  `json:"speedText"`    // 格式化的速度文本
}
