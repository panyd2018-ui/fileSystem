package handlers

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filesystem/internal/config"
	"filesystem/internal/models"
	"filesystem/internal/utils"

	"github.com/gorilla/mux"
)

var staticFiles embed.FS

// InitHandlers 初始化 handlers，设置静态文件
func InitHandlers(fs embed.FS) {
	staticFiles = fs
}

// ServeIndex 返回前端页面
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INDEX] 请求开始 - 方法: %s, 路径: %s, 客户端IP: %s, User-Agent: %s",
		r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

	// 从嵌入的文件系统读取
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Printf("[INDEX] 错误: 无法读取嵌入的静态文件系统 - %v", err)
		http.Error(w, "无法读取前端文件", http.StatusInternalServerError)
		return
	}

	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		log.Printf("[INDEX] 错误: 无法打开 index.html - %v", err)
		http.Error(w, "无法读取前端页面", http.StatusInternalServerError)
		return
	}
	defer indexFile.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	bytesWritten, err := io.Copy(w, indexFile)
	if err != nil {
		log.Printf("[INDEX] 错误: 写入响应失败 - %v", err)
		return
	}
	log.Printf("[INDEX] 成功: 已返回前端页面, 大小: %d 字节", bytesWritten)
}

// ListFiles 列出所有文件
func ListFiles(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	path := r.URL.Query().Get("path")
	log.Printf("[LIST] 请求开始 - 方法: %s, 路径参数: %s, 客户端IP: %s, User-Agent: %s",
		r.Method, path, r.RemoteAddr, r.UserAgent())

	// 验证路径
	if path != "" {
		if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			log.Printf("[LIST] 错误: 无效的路径参数 - path=%s (包含 '..' 或以 '/' 开头)", path)
			utils.SendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
	}

	// 构建完整路径
	targetDir := config.UploadDir
	if path != "" {
		targetDir = filepath.Join(config.UploadDir, path)
		absTarget, _ := filepath.Abs(targetDir)
		absUpload, _ := filepath.Abs(config.UploadDir)
		log.Printf("[LIST] 路径验证 - 目标目录: %s, 绝对路径: %s, 基础目录: %s",
			targetDir, absTarget, absUpload)
		if !strings.HasPrefix(absTarget, absUpload) {
			log.Printf("[LIST] 错误: 路径遍历攻击尝试 - 目标: %s, 基础: %s", absTarget, absUpload)
			utils.SendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
	}

	log.Printf("[LIST] 正在读取目录: %s", targetDir)
	files, err := os.ReadDir(targetDir)
	if err != nil {
		log.Printf("[LIST] 错误: 无法读取目录 %s - %v", targetDir, err)
		utils.SendError(w, "无法读取文件列表", http.StatusInternalServerError)
		return
	}

	log.Printf("[LIST] 找到 %d 个文件/目录", len(files))
	var fileList []models.FileInfo
	skippedCount := 0
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			log.Printf("[LIST] 警告: 无法获取文件信息 %s - %v", file.Name(), err)
			skippedCount++
			continue
		}

		relativePath := file.Name()
		if path != "" {
			relativePath = filepath.Join(path, file.Name())
		}

		fileInfo := models.FileInfo{
			Name:      file.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsDir:     file.IsDir(),
			Extension: strings.TrimPrefix(filepath.Ext(file.Name()), "."),
			Path:      relativePath,
		}
		fileList = append(fileList, fileInfo)
	}

	if skippedCount > 0 {
		log.Printf("[LIST] 警告: 跳过了 %d 个无法读取的文件", skippedCount)
	}

	duration := time.Since(startTime)
	log.Printf("[LIST] 成功: 返回 %d 个文件/目录, 耗时: %v", len(fileList), duration)
	utils.SendJSON(w, models.Response{
		Success: true,
		Data:    fileList,
	})
}

// UploadFile 上传文件
func UploadFile(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	uploadPath := r.URL.Query().Get("path")
	log.Printf("[UPLOAD] 请求开始 - 方法: %s, 上传路径参数: %s, 客户端IP: %s, User-Agent: %s, Content-Type: %s",
		r.Method, uploadPath, r.RemoteAddr, r.UserAgent(), r.Header.Get("Content-Type"))

	// 验证路径
	if uploadPath != "" {
		if strings.Contains(uploadPath, "..") || strings.HasPrefix(uploadPath, "/") {
			log.Printf("[UPLOAD] 错误: 无效的上传路径参数 - path=%s", uploadPath)
			utils.SendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
	}

	// 解析 multipart form
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Printf("[UPLOAD] 标准解析失败，尝试流式处理 - 错误: %v", err)
		handleStreamUpload(w, r, uploadPath, startTime)
		return
	}

	// 标准方式处理（小文件）
	handleStandardUpload(w, r, uploadPath, startTime)
}

// 处理流式上传（大文件）
func handleStreamUpload(w http.ResponseWriter, r *http.Request, uploadPath string, startTime time.Time) {
	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("[UPLOAD] 错误: 无法创建 MultipartReader - %v", err)
		utils.SendError(w, "无法解析上传请求", http.StatusBadRequest)
		return
	}

	part, err := reader.NextPart()
	if err != nil {
		log.Printf("[UPLOAD] 错误: 无法读取文件 part - %v", err)
		utils.SendError(w, "无法读取文件数据", http.StatusBadRequest)
		return
	}
	defer part.Close()

	filename := part.FileName()
	log.Printf("[UPLOAD] 流式上传 - 文件名: %s, Content-Type: %s", filename, part.Header.Get("Content-Type"))
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		log.Printf("[UPLOAD] 错误: 无效的文件名 - filename=%s", filename)
		utils.SendError(w, "无效的文件名", http.StatusBadRequest)
		return
	}

	targetPath, err := validateAndPreparePath(uploadPath)
	if err != nil {
		log.Printf("[UPLOAD] 错误: 路径验证失败 - %v", err)
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(targetPath, filename)
	log.Printf("[UPLOAD] 正在创建目标文件: %s", fullPath)
	dst, err := os.Create(fullPath)
	if err != nil {
		log.Printf("[UPLOAD] 错误: 无法创建文件 %s - %v", fullPath, err)
		utils.SendError(w, "无法创建文件", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 使用速度跟踪器
	speedTracker := utils.NewSpeedTracker(dst)
	log.Printf("[UPLOAD] 开始流式复制文件内容...")

	// 在后台定期打印速度
	stopSpeedLog := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				currentSpeed := speedTracker.GetSpeed()
				avgSpeed := speedTracker.GetAverageSpeed()
				totalBytes := speedTracker.GetTotalBytes()
				if currentSpeed > 0 {
					log.Printf("[UPLOAD] 传输中 - 已传输: %s, 当前速度: %s, 平均速度: %s",
						utils.FormatSize(totalBytes), utils.FormatSpeed(currentSpeed), utils.FormatSpeed(avgSpeed))
				}
			case <-stopSpeedLog:
				return
			}
		}
	}()

	bytesWritten, err := io.Copy(speedTracker, part)
	stopSpeedLog <- true

	if err != nil {
		log.Printf("[UPLOAD] 错误: 文件写入失败 - 文件: %s, 已写入: %d 字节, 错误: %v",
			fullPath, bytesWritten, err)
		os.Remove(dst.Name())
		log.Printf("[UPLOAD] 已删除不完整的文件: %s", dst.Name())
		utils.SendError(w, "无法保存文件", http.StatusInternalServerError)
		return
	}

	duration := time.Since(startTime)
	avgSpeed := speedTracker.GetAverageSpeed()

	log.Printf("[UPLOAD] 成功: 文件 %s 上传完成, 大小: %s, 耗时: %v, 平均速度: %s",
		filename, utils.FormatSize(bytesWritten), duration, utils.FormatSpeed(avgSpeed))

	utils.SendJSON(w, models.Response{
		Success: true,
		Message: fmt.Sprintf("文件 %s 上传成功", filename),
		Speed: &models.SpeedInfo{
			AverageSpeed: avgSpeed,
			CurrentSpeed: speedTracker.GetSpeed(),
			TotalBytes:   bytesWritten,
			Duration:     duration.String(),
			SpeedText:    utils.FormatSpeed(avgSpeed),
		},
	})
}

// 处理标准上传（小文件）
func handleStandardUpload(w http.ResponseWriter, r *http.Request, uploadPath string, startTime time.Time) {
	log.Printf("[UPLOAD] 使用标准方式处理（小文件）")
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("[UPLOAD] 错误: 无法获取上传的文件 - %v", err)
		utils.SendError(w, "无法获取上传的文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := handler.Filename
	log.Printf("[UPLOAD] 标准上传 - 文件名: %s, 大小: %s, Content-Type: %s",
		filename, utils.FormatSize(handler.Size), handler.Header.Get("Content-Type"))
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		log.Printf("[UPLOAD] 错误: 无效的文件名 - filename=%s", filename)
		utils.SendError(w, "无效的文件名", http.StatusBadRequest)
		return
	}

	targetPath, err := validateAndPreparePath(uploadPath)
	if err != nil {
		log.Printf("[UPLOAD] 错误: 路径验证失败 - %v", err)
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(targetPath, filename)
	log.Printf("[UPLOAD] 正在创建目标文件: %s", fullPath)
	dst, err := os.Create(fullPath)
	if err != nil {
		log.Printf("[UPLOAD] 错误: 无法创建文件 %s - %v", fullPath, err)
		utils.SendError(w, "无法创建文件", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 使用速度跟踪器
	speedTracker := utils.NewSpeedTracker(dst)
	log.Printf("[UPLOAD] 开始复制文件内容...")

	bytesWritten, err := io.Copy(speedTracker, file)
	if err != nil {
		log.Printf("[UPLOAD] 错误: 文件写入失败 - 文件: %s, 已写入: %d 字节, 错误: %v",
			fullPath, bytesWritten, err)
		os.Remove(dst.Name())
		log.Printf("[UPLOAD] 已删除不完整的文件: %s", dst.Name())
		utils.SendError(w, "无法保存文件", http.StatusInternalServerError)
		return
	}

	fileInfo, _ := os.Stat(fullPath)
	duration := time.Since(startTime)
	avgSpeed := speedTracker.GetAverageSpeed()

	log.Printf("[UPLOAD] 成功: 文件 %s 上传完成, 大小: %s, 耗时: %v, 平均速度: %s",
		filename, utils.FormatSize(bytesWritten), duration, utils.FormatSpeed(avgSpeed))
	if fileInfo != nil {
		log.Printf("[UPLOAD] 文件信息 - 最终大小: %d 字节, 修改时间: %v",
			fileInfo.Size(), fileInfo.ModTime())
	}

	utils.SendJSON(w, models.Response{
		Success: true,
		Message: fmt.Sprintf("文件 %s 上传成功", filename),
		Speed: &models.SpeedInfo{
			AverageSpeed: avgSpeed,
			CurrentSpeed: speedTracker.GetSpeed(),
			TotalBytes:   bytesWritten,
			Duration:     duration.String(),
			SpeedText:    utils.FormatSpeed(avgSpeed),
		},
	})
}

// validateAndPreparePath 验证并准备路径
func validateAndPreparePath(uploadPath string) (string, error) {
	targetPath := config.UploadDir
	if uploadPath != "" {
		targetPath = filepath.Join(config.UploadDir, uploadPath)
		absTarget, _ := filepath.Abs(targetPath)
		absUpload, _ := filepath.Abs(config.UploadDir)
		log.Printf("[PATH] 路径验证 - 目标路径: %s, 绝对路径: %s, 基础目录: %s",
			targetPath, absTarget, absUpload)
		if !strings.HasPrefix(absTarget, absUpload) {
			log.Printf("[PATH] 错误: 路径遍历攻击尝试 - 目标: %s, 基础: %s", absTarget, absUpload)
			return "", fmt.Errorf("无效的路径")
		}
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			log.Printf("[PATH] 错误: 无法创建目录 %s - %v", targetPath, err)
			return "", fmt.Errorf("无法创建目录")
		}
		log.Printf("[PATH] 已创建/确认目录存在: %s", targetPath)
	}
	return targetPath, nil
}

// DownloadFile 下载文件
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	vars := mux.Vars(r)
	filePath := vars["filename"]
	log.Printf("[DOWNLOAD] 请求开始 - 方法: %s, 文件路径: %s, 客户端IP: %s, User-Agent: %s",
		r.Method, filePath, r.RemoteAddr, r.UserAgent())

	// 验证路径
	if filePath == "" || strings.Contains(filePath, "..") {
		log.Printf("[DOWNLOAD] 错误: 无效的文件路径 - filePath=%s", filePath)
		utils.SendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(config.UploadDir, filePath)
	log.Printf("[DOWNLOAD] 构建完整路径 - 基础目录: %s, 相对路径: %s, 完整路径: %s",
		config.UploadDir, filePath, fullPath)

	// 确保路径在 config.UploadDir 内
	absTarget, _ := filepath.Abs(fullPath)
	absUpload, _ := filepath.Abs(config.UploadDir)
	log.Printf("[DOWNLOAD] 路径验证 - 绝对目标路径: %s, 绝对基础路径: %s", absTarget, absUpload)
	if !strings.HasPrefix(absTarget, absUpload) {
		log.Printf("[DOWNLOAD] 错误: 路径遍历攻击尝试 - 目标: %s, 基础: %s", absTarget, absUpload)
		utils.SendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	log.Printf("[DOWNLOAD] 正在检查文件是否存在: %s", fullPath)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		log.Printf("[DOWNLOAD] 错误: 文件不存在 - %s, 错误: %v", fullPath, err)
		utils.SendError(w, "文件不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("[DOWNLOAD] 错误: 无法获取文件信息 - %s, 错误: %v", fullPath, err)
		utils.SendError(w, "无法访问文件", http.StatusInternalServerError)
		return
	}

	log.Printf("[DOWNLOAD] 文件信息 - 名称: %s, 大小: %s, 是否为目录: %v, 修改时间: %v",
		info.Name(), utils.FormatSize(info.Size()), info.IsDir(), info.ModTime())

	// 如果是目录，不允许下载
	if info.IsDir() {
		log.Printf("[DOWNLOAD] 错误: 尝试下载目录 - %s", fullPath)
		utils.SendError(w, "不能下载目录", http.StatusBadRequest)
		return
	}

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("[DOWNLOAD] 错误: 无法打开文件 - %s, 错误: %v", fullPath, err)
		utils.SendError(w, "无法打开文件", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 设置响应头
	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	log.Printf("[DOWNLOAD] 已设置响应头 - Content-Disposition: attachment; filename=%s, 大小: %s",
		filename, utils.FormatSize(info.Size()))

	// 使用速度跟踪器
	speedTracker := utils.NewSpeedTrackerReader(file)

	// 在后台定期打印速度
	stopSpeedLog := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				currentSpeed := speedTracker.GetSpeed()
				avgSpeed := speedTracker.GetAverageSpeed()
				totalBytes := speedTracker.GetTotalBytes()
				if currentSpeed > 0 {
					log.Printf("[DOWNLOAD] 传输中 - 已传输: %s/%s, 当前速度: %s, 平均速度: %s",
						utils.FormatSize(totalBytes), utils.FormatSize(info.Size()), utils.FormatSpeed(currentSpeed), utils.FormatSpeed(avgSpeed))
				}
			case <-stopSpeedLog:
				return
			}
		}
	}()

	// 发送文件
	log.Printf("[DOWNLOAD] 开始发送文件内容...")
	bytesWritten, err := io.Copy(w, speedTracker)
	stopSpeedLog <- true

	if err != nil {
		log.Printf("[DOWNLOAD] 错误: 文件传输失败 - 文件: %s, 已传输: %d 字节, 错误: %v",
			fullPath, bytesWritten, err)
		return
	}

	duration := time.Since(startTime)
	avgSpeed := speedTracker.GetAverageSpeed()
	log.Printf("[DOWNLOAD] 成功: 文件 %s 下载完成, 大小: %s, 耗时: %v, 平均速度: %s",
		filename, utils.FormatSize(bytesWritten), duration, utils.FormatSpeed(avgSpeed))
}

// DeleteFile 删除文件
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	vars := mux.Vars(r)
	filePath := vars["filename"]
	log.Printf("[DELETE] 请求开始 - 方法: %s, 文件路径: %s, 客户端IP: %s, User-Agent: %s",
		r.Method, filePath, r.RemoteAddr, r.UserAgent())

	// 验证路径
	if filePath == "" || strings.Contains(filePath, "..") {
		log.Printf("[DELETE] 错误: 无效的文件路径 - filePath=%s", filePath)
		utils.SendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(config.UploadDir, filePath)
	log.Printf("[DELETE] 构建完整路径 - 基础目录: %s, 相对路径: %s, 完整路径: %s",
		config.UploadDir, filePath, fullPath)

	// 确保路径在 config.UploadDir 内
	absTarget, _ := filepath.Abs(fullPath)
	absUpload, _ := filepath.Abs(config.UploadDir)
	log.Printf("[DELETE] 路径验证 - 绝对目标路径: %s, 绝对基础路径: %s", absTarget, absUpload)
	if !strings.HasPrefix(absTarget, absUpload) {
		log.Printf("[DELETE] 错误: 路径遍历攻击尝试 - 目标: %s, 基础: %s", absTarget, absUpload)
		utils.SendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	log.Printf("[DELETE] 正在检查文件/目录是否存在: %s", fullPath)
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		log.Printf("[DELETE] 错误: 文件或目录不存在 - %s, 错误: %v", fullPath, err)
		utils.SendError(w, "文件或目录不存在", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("[DELETE] 错误: 无法获取文件信息 - %s, 错误: %v", fullPath, err)
		utils.SendError(w, "无法访问文件", http.StatusInternalServerError)
		return
	}

	itemType := "目录"
	if !info.IsDir() {
		itemType = "文件"
	}
	log.Printf("[DELETE] 准备删除 %s - 名称: %s, 大小: %s, 修改时间: %v",
		itemType, info.Name(), utils.FormatSize(info.Size()), info.ModTime())

	// 删除文件或目录
	var err2 error
	if info.IsDir() {
		log.Printf("[DELETE] 正在删除目录及其所有内容: %s", fullPath)
		err2 = os.RemoveAll(fullPath)
	} else {
		log.Printf("[DELETE] 正在删除文件: %s", fullPath)
		err2 = os.Remove(fullPath)
	}

	if err2 != nil {
		log.Printf("[DELETE] 错误: 删除失败 - %s, 错误: %v", fullPath, err2)
		utils.SendError(w, "无法删除", http.StatusInternalServerError)
		return
	}

	name := filepath.Base(filePath)
	duration := time.Since(startTime)
	log.Printf("[DELETE] 成功: %s %s 删除完成, 耗时: %v", itemType, name, duration)

	utils.SendJSON(w, models.Response{
		Success: true,
		Message: fmt.Sprintf("%s %s 删除成功", itemType, name),
	})
}
