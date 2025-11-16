package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SendJSON 发送 JSON 响应
func SendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// SendError 发送错误响应
func SendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// SpeedTracker 用于跟踪传输速度
type SpeedTracker struct {
	writer     io.Writer
	reader     io.Reader
	startTime  time.Time
	totalBytes int64
	lastBytes  int64
	lastTime   time.Time
}

// NewSpeedTracker 创建速度跟踪器
func NewSpeedTracker(w io.Writer) *SpeedTracker {
	now := time.Now()
	return &SpeedTracker{
		writer:    w,
		startTime: now,
		lastTime:  now,
	}
}

// NewSpeedTrackerReader 创建读取速度跟踪器
func NewSpeedTrackerReader(r io.Reader) *SpeedTracker {
	now := time.Now()
	return &SpeedTracker{
		reader:    r,
		startTime: now,
		lastTime:  now,
	}
}

// Write 实现 io.Writer 接口
func (st *SpeedTracker) Write(p []byte) (n int, err error) {
	n, err = st.writer.Write(p)
	if n > 0 {
		st.totalBytes += int64(n)
	}
	return
}

// Read 实现 io.Reader 接口
func (st *SpeedTracker) Read(p []byte) (n int, err error) {
	n, err = st.reader.Read(p)
	if n > 0 {
		st.totalBytes += int64(n)
	}
	return
}

// GetSpeed 获取当前速度（字节/秒）
func (st *SpeedTracker) GetSpeed() float64 {
	now := time.Now()
	elapsed := now.Sub(st.lastTime).Seconds()
	if elapsed < 0.1 { // 至少间隔100ms才计算速度
		return 0
	}

	bytesDelta := st.totalBytes - st.lastBytes
	speed := float64(bytesDelta) / elapsed

	st.lastBytes = st.totalBytes
	st.lastTime = now

	return speed
}

// GetAverageSpeed 获取平均速度（字节/秒）
func (st *SpeedTracker) GetAverageSpeed() float64 {
	elapsed := time.Since(st.startTime).Seconds()
	if elapsed < 0.1 {
		return 0
	}
	return float64(st.totalBytes) / elapsed
}

// GetTotalBytes 获取总传输字节数
func (st *SpeedTracker) GetTotalBytes() int64 {
	return st.totalBytes
}

// FormatSpeed 格式化速度显示
func FormatSpeed(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
	}
}

// FormatSize 格式化文件大小显示
func FormatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
	}
}
