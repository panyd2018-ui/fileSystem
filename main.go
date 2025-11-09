package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

//go:embed static
var staticFiles embed.FS

func init() {
	loadConfig()
}

func main() {
	r := setupRoutes()

	absPath, _ := filepath.Abs(uploadDir)
	fmt.Printf("文件系统服务器启动在 http://localhost%s\n", port)
	fmt.Printf("存储目录: %s\n", absPath)
	fmt.Printf("提示: 可通过 config.json 配置文件修改存储目录和端口\n")

	log.Fatal(http.ListenAndServe(port, corsMiddleware(r)))
}

// setupRoutes 设置路由
func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// 静态文件服务（从嵌入的文件系统读取）
	// 使用 fs.Sub 获取 static 子目录
	staticSubFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("无法读取嵌入的静态文件: %v", err)
	}
	staticFS := http.FS(staticSubFS)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(staticFS)))

	// API 路由
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/files", listFiles).Methods("GET")
	api.HandleFunc("/upload", uploadFile).Methods("POST")
	api.HandleFunc("/download/{filename}", downloadFile).Methods("GET")
	api.HandleFunc("/delete/{filename}", deleteFile).Methods("DELETE")

	// 前端页面
	r.HandleFunc("/", serveIndex).Methods("GET")

	return r
}
