package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// 返回前端页面
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

// 列出所有文件
func listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(uploadDir)
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

		fileInfo := FileInfo{
			Name:      file.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsDir:     file.IsDir(),
			Extension: strings.TrimPrefix(filepath.Ext(file.Name()), "."),
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

		// 创建目标文件
		dst, err := os.Create(filepath.Join(uploadDir, filename))
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

	// 创建目标文件
	dst, err := os.Create(filepath.Join(uploadDir, filename))
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
	filename := vars["filename"]

	// 验证文件名
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		sendError(w, "无效的文件名", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		sendError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// 发送文件
	http.ServeFile(w, r, filePath)
}

// 删除文件
func deleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	// 验证文件名
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		sendError(w, "无效的文件名", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		sendError(w, "文件不存在", http.StatusNotFound)
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		sendError(w, "无法删除文件", http.StatusInternalServerError)
		return
	}

	sendJSON(w, Response{
		Success: true,
		Message: fmt.Sprintf("文件 %s 删除成功", filename),
	})
}
