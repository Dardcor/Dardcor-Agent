package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	"dardcor-agent/config"
	"dardcor-agent/handlers"
	"dardcor-agent/middleware"
	"dardcor-agent/services"
	"dardcor-agent/storage"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize configuration
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Initialize storage
	storage.Init()

	// Initialize services
	fsSvc := services.NewFileSystemService()
	cmdSvc := services.NewCommandService()
	sysSvc := services.NewSystemService()
	agentSvc := services.NewAgentService(fsSvc, cmdSvc, sysSvc)

	// Initialize handlers
	fsHandler := handlers.NewFileSystemHandler(fsSvc)
	cmdHandler := handlers.NewCommandHandler(cmdSvc)
	sysHandler := handlers.NewSystemHandler(sysSvc)
	agentHandler := handlers.NewAgentHandler(agentSvc)
	wsHandler := handlers.NewWebSocketHandler(agentSvc, cmdSvc)

	// Setup router
	r := mux.NewRouter()

	// API Routes
	api := r.PathPrefix("/api").Subrouter()

	// Agent routes
	api.HandleFunc("/agent", agentHandler.ProcessMessage).Methods("POST", "OPTIONS")

	// File system routes
	api.HandleFunc("/files", fsHandler.ListDirectory).Methods("GET")
	api.HandleFunc("/files", fsHandler.DeleteFile).Methods("DELETE")
	api.HandleFunc("/files/read", fsHandler.ReadFile).Methods("GET")
	api.HandleFunc("/files/write", fsHandler.WriteFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/files/search", fsHandler.SearchFiles).Methods("POST", "OPTIONS")
	api.HandleFunc("/files/info", fsHandler.GetFileInfo).Methods("GET")
	api.HandleFunc("/files/mkdir", fsHandler.CreateDirectory).Methods("POST", "OPTIONS")
	api.HandleFunc("/files/drives", fsHandler.GetDrives).Methods("GET")
	api.HandleFunc("/files/move", fsHandler.MoveFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/files/copy", fsHandler.CopyFile).Methods("POST", "OPTIONS")

	// Command routes
	api.HandleFunc("/command", cmdHandler.ExecuteCommand).Methods("POST", "OPTIONS")
	api.HandleFunc("/command/history", cmdHandler.GetCommandHistory).Methods("GET")
	api.HandleFunc("/command/info", cmdHandler.GetShellInfo).Methods("GET")

	// System routes
	api.HandleFunc("/system", sysHandler.GetSystemInfo).Methods("GET")
	api.HandleFunc("/system/processes", sysHandler.GetProcesses).Methods("GET")
	api.HandleFunc("/system/kill", sysHandler.KillProcess).Methods("POST", "OPTIONS")
	api.HandleFunc("/system/cpu", sysHandler.GetCPUUsage).Methods("GET")
	api.HandleFunc("/system/memory", sysHandler.GetMemoryUsage).Methods("GET")

	// WebSocket route
	r.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Serve frontend static files
	frontendDevUrl := os.Getenv("DARDCOR_DEV_URL")
	if frontendDevUrl != "" {
		log.Printf("🚀 Development Mode: Proxying frontend internally to %s", frontendDevUrl)
		
		// Setup proxy
		target, _ := url.Parse(frontendDevUrl)
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Don't proxy if it's already handled by API or WebSocket
			proxy.ServeHTTP(w, req)
		})
	} else {
		frontendDist := filepath.Join("dist")
		if _, err := os.Stat(frontendDist); err == nil {
			spa := spaHandler{staticPath: frontendDist, indexPath: "index.html"}
			r.PathPrefix("/").Handler(spa)
		} else {
			// Development mode - serve a simple health check
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"status":"running","name":"Dardcor Agent","version":"1.0.0","port":"%s"}`, cfg.Port)
			})
		}
	}

	// Apply middleware
	handler := middleware.CORS(middleware.Logger(r))

	// Start server
	addr := "127.0.0.1:" + cfg.Port
	log.Printf("🤖 Dardcor Agent Server starting on http://%s", addr)
	log.Printf("📁 Data directory: %s", cfg.DataDir)
	log.Printf("🔌 WebSocket endpoint: ws://%s/ws", addr)
	log.Printf("📡 API endpoint: http://%s/api", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// spaHandler serves the frontend SPA
type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticPath, r.URL.Path)

	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Serve index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
