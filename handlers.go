package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// 返回前端页面
func serveIndex(w http.ResponseWriter, r *http.Request) {
	// 从嵌入的文件系统读取
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		http.Error(w, "无法读取前端文件", http.StatusInternalServerError)
		return
	}

	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		http.Error(w, "无法读取前端页面", http.StatusInternalServerError)
		return
	}
	defer indexFile.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, indexFile)
}

// 列出所有文件
func listFiles(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path != "" {
		// 验证路径，防止路径遍历攻击
		if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
			sendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
	}

	// 构建完整路径
	targetDir := uploadDir
	if path != "" {
		targetDir = filepath.Join(uploadDir, path)
		// 确保路径在 uploadDir 内
		absTarget, _ := filepath.Abs(targetDir)
		absUpload, _ := filepath.Abs(uploadDir)
		if !strings.HasPrefix(absTarget, absUpload) {
			sendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
	}

	files, err := os.ReadDir(targetDir)
	if err != nil {
		sendError(w, "无法读取文件列表", http.StatusInternalServerError)
		return
	}

	var fileList []FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		// 构建相对路径
		relativePath := file.Name()
		if path != "" {
			relativePath = filepath.Join(path, file.Name())
		}

		fileInfo := FileInfo{
			Name:      file.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsDir:     file.IsDir(),
			Extension: strings.TrimPrefix(filepath.Ext(file.Name()), "."),
			Path:      relativePath,
		}
		fileList = append(fileList, fileInfo)
	}

	sendJSON(w, Response{
		Success: true,
		Data:    fileList,
	})
}

// 上传文件
func uploadFile(w http.ResponseWriter, r *http.Request) {
	// 获取上传路径参数
	uploadPath := r.URL.Query().Get("path")
	if uploadPath != "" {
		// 验证路径，防止路径遍历攻击
		if strings.Contains(uploadPath, "..") || strings.HasPrefix(uploadPath, "/") {
			sendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
	}

	// 解析 multipart form，不限制文件大小
	// 32MB 只是内存缓冲区大小，实际文件会流式写入磁盘
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		// 如果解析失败，尝试直接读取文件流
		// 这允许上传超大文件
		reader, err := r.MultipartReader()
		if err != nil {
			sendError(w, "无法解析上传请求", http.StatusBadRequest)
			return
		}

		// 读取第一个 part
		part, err := reader.NextPart()
		if err != nil {
			sendError(w, "无法读取文件数据", http.StatusBadRequest)
			return
		}
		defer part.Close()

		// 获取文件名
		filename := part.FileName()
		if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
			sendError(w, "无效的文件名", http.StatusBadRequest)
			return
		}

		// 构建目标路径
		targetPath := uploadDir
		if uploadPath != "" {
			targetPath = filepath.Join(uploadDir, uploadPath)
			// 确保路径在 uploadDir 内
			absTarget, _ := filepath.Abs(targetPath)
			absUpload, _ := filepath.Abs(uploadDir)
			if !strings.HasPrefix(absTarget, absUpload) {
				sendError(w, "无效的路径", http.StatusBadRequest)
				return
			}
			// 确保目录存在
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				sendError(w, "无法创建目录", http.StatusInternalServerError)
				return
			}
		}

		// 创建目标文件
		dst, err := os.Create(filepath.Join(targetPath, filename))
		if err != nil {
			sendError(w, "无法创建文件", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// 流式复制文件内容
		if _, err := io.Copy(dst, part); err != nil {
			os.Remove(dst.Name()) // 删除不完整的文件
			sendError(w, "无法保存文件", http.StatusInternalServerError)
			return
		}

		sendJSON(w, Response{
			Success: true,
			Message: fmt.Sprintf("文件 %s 上传成功", filename),
		})
		return
	}

	// 标准方式处理（小文件）
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendError(w, "无法获取上传的文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件名
	filename := handler.Filename
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		sendError(w, "无效的文件名", http.StatusBadRequest)
		return
	}

	// 构建目标路径
	targetPath := uploadDir
	if uploadPath != "" {
		targetPath = filepath.Join(uploadDir, uploadPath)
		// 确保路径在 uploadDir 内
		absTarget, _ := filepath.Abs(targetPath)
		absUpload, _ := filepath.Abs(uploadDir)
		if !strings.HasPrefix(absTarget, absUpload) {
			sendError(w, "无效的路径", http.StatusBadRequest)
			return
		}
		// 确保目录存在
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			sendError(w, "无法创建目录", http.StatusInternalServerError)
			return
		}
	}

	// 创建目标文件
	dst, err := os.Create(filepath.Join(targetPath, filename))
	if err != nil {
		sendError(w, "无法创建文件", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(dst.Name()) // 删除不完整的文件
		sendError(w, "无法保存文件", http.StatusInternalServerError)
		return
	}

	sendJSON(w, Response{
		Success: true,
		Message: fmt.Sprintf("文件 %s 上传成功", filename),
	})
}

// 下载文件
func downloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["filename"]

	// 验证路径
	if filePath == "" || strings.Contains(filePath, "..") {
		sendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(uploadDir, filePath)

	// 确保路径在 uploadDir 内
	absTarget, _ := filepath.Abs(fullPath)
	absUpload, _ := filepath.Abs(uploadDir)
	if !strings.HasPrefix(absTarget, absUpload) {
		sendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		sendError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 如果是目录，不允许下载
	if info.IsDir() {
		sendError(w, "不能下载目录", http.StatusBadRequest)
		return
	}

	// 获取文件名
	filename := filepath.Base(filePath)

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// 发送文件
	http.ServeFile(w, r, fullPath)
}

// 删除文件
func deleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["filename"]

	// 验证路径
	if filePath == "" || strings.Contains(filePath, "..") {
		sendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(uploadDir, filePath)

	// 确保路径在 uploadDir 内
	absTarget, _ := filepath.Abs(fullPath)
	absUpload, _ := filepath.Abs(uploadDir)
	if !strings.HasPrefix(absTarget, absUpload) {
		sendError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		sendError(w, "文件或目录不存在", http.StatusNotFound)
		return
	}

	// 删除文件或目录
	var err2 error
	if info.IsDir() {
		err2 = os.RemoveAll(fullPath)
	} else {
		err2 = os.Remove(fullPath)
	}

	if err2 != nil {
		sendError(w, "无法删除", http.StatusInternalServerError)
		return
	}

	name := filepath.Base(filePath)
	itemType := "目录"
	if !info.IsDir() {
		itemType = "文件"
	}

	sendJSON(w, Response{
		Success: true,
		Message: fmt.Sprintf("%s %s 删除成功", itemType, name),
	})
}
