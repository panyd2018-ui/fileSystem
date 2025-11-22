package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"fileSystem/internal/config"
	"fileSystem/internal/handlers"
	"fileSystem/internal/middleware"

	"github.com/gorilla/mux"
)

//go:embed static
var staticFiles embed.FS

func init() {
	config.LoadConfig()
	handlers.InitHandlers(staticFiles)
}

func main() {
	r := setupRoutes()

	absPath, _ := filepath.Abs(config.UploadDir)
	fmt.Printf("文件系统服务器启动在 http://localhost%s\n", config.Port)
	fmt.Printf("存储目录: %s\n", absPath)
	fmt.Printf("提示: 可通过 config.json 配置文件修改存储目录、端口和根路由\n")

	log.Fatal(http.ListenAndServe(config.Port, middleware.CORSMiddleware(r)))
}

// 规范化根路径
func normalizeRootPath(path string) string {
	if path == "" {
		return "/"
	}
	// 确保以 / 开头
	if path[0] != '/' {
		path = "/" + path
	}
	// 如果根路由不是 "/"，去掉末尾的 "/"（如果有）
	if path != "/" && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

// setupRoutes 设置路由
func setupRoutes() *mux.Router {
	r := mux.NewRouter()

	// 规范化根路径
	rootPath := normalizeRootPath(config.Cfg.RootPath)

	// 静态文件服务（从嵌入的文件系统读取）
	// 使用 fs.Sub 获取 static 子目录
	staticSubFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("无法读取嵌入的静态文件: %v", err)
	}
	staticFS := http.FS(staticSubFS)

	// 创建静态文件处理器，添加日志
	loggedStaticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[STATIC] 请求静态文件 - 路径: %s, 客户端IP: %s", r.URL.Path, r.RemoteAddr)

		// 根据根路径选择正确的处理器
		var handler http.Handler
		if rootPath == "/" {
			handler = http.StripPrefix("/static/", http.FileServer(staticFS))
		} else {
			handler = http.StripPrefix(rootPath+"/static/", http.FileServer(staticFS))
		}
		handler.ServeHTTP(w, r)
	})

	// 静态文件路由 - 使用根路径前缀
	if rootPath == "/" {
		r.PathPrefix("/static/").Handler(loggedStaticHandler)
	} else {
		r.PathPrefix(rootPath + "/static/").Handler(loggedStaticHandler)
	}

	// API 路由 - 使用根路径前缀
	var api *mux.Router
	if rootPath == "/" {
		api = r.PathPrefix("/api").Subrouter()
	} else {
		api = r.PathPrefix(rootPath + "/api").Subrouter()
	}
	api.HandleFunc("/files", handlers.ListFiles).Methods("GET")
	api.HandleFunc("/upload", handlers.UploadFile).Methods("POST")
	api.HandleFunc("/download/{filename:.*}", handlers.DownloadFile).Methods("GET")
	api.HandleFunc("/delete/{filename:.*}", handlers.DeleteFile).Methods("DELETE")

	// 前端页面 - 使用配置的根路由
	r.HandleFunc(rootPath, handlers.ServeIndex).Methods("GET")

	return r
}
