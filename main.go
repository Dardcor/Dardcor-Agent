package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/handlers"
	"dardcor-agent/middleware"
	"dardcor-agent/services"
	"dardcor-agent/storage"

	"github.com/gorilla/mux"
)

var startTime = time.Now()

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	storage.Init()

	fsSvc := services.NewFileSystemService()
	cmdSvc := services.NewCommandService()
	sysSvc := services.NewSystemService()
	agentSvc := services.NewAgentService(fsSvc, cmdSvc, sysSvc)

	fsHandler := handlers.NewFileSystemHandler(fsSvc)
	cmdHandler := handlers.NewCommandHandler(cmdSvc)
	sysHandler := handlers.NewSystemHandler(sysSvc)
	agentHandler := handlers.NewAgentHandler(agentSvc)
	wsHandler := handlers.NewWebSocketHandler(agentSvc, cmdSvc)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/agent", agentHandler.ProcessMessage).Methods("POST", "OPTIONS")

	r.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"uptime":   time.Since(startTime).String(),
			"provider": cfg.AI.Provider,
			"model":    cfg.AI.Model,
		})
	}).Methods("GET")

	api.HandleFunc("/system", sysHandler.GetSystemInfo).Methods("GET")

	api.HandleFunc("/provider", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		hostname, _ := os.Hostname()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":     "Dardcor Agent",
			"provider": cfg.AI.Provider,
			"model":    cfg.AI.Model,
			"hostname": hostname,
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"uptime":   time.Since(startTime).String(),
			"port":     cfg.Port,
		})
	}).Methods("GET")

	api.HandleFunc("/skills", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		skills := []map[string]string{
			{"name": "list_directory", "desc": "List files in a directory", "cmd": "list <path>"},
			{"name": "read_file", "desc": "Read file contents", "cmd": "read <path>"},
			{"name": "write_file", "desc": "Write to a file", "cmd": "write <path> <content>"},
			{"name": "delete_file", "desc": "Delete a file or folder", "cmd": "delete <path>"},
			{"name": "search_files", "desc": "Search for files", "cmd": "search <query>"},
			{"name": "create_directory", "desc": "Create a directory", "cmd": "mkdir <path>"},
			{"name": "execute_command", "desc": "Execute a shell command", "cmd": "run <command>"},
			{"name": "system_info", "desc": "Get system information", "cmd": "sysinfo"},
			{"name": "list_processes", "desc": "List running processes", "cmd": "processes"},
			{"name": "kill_process", "desc": "Kill a process by PID", "cmd": "kill <pid>"},
			{"name": "cpu_info", "desc": "Get CPU information", "cmd": "cpu"},
			{"name": "memory_info", "desc": "Get memory information", "cmd": "memory"},
			{"name": "list_drives", "desc": "List available drives", "cmd": "drives"},
			{"name": "file_info", "desc": "Get file/directory info", "cmd": "info <path>"},
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    skills,
			"count":   len(skills),
		})
	}).Methods("GET")

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

	api.HandleFunc("/command", cmdHandler.ExecuteCommand).Methods("POST", "OPTIONS")
	api.HandleFunc("/command/history", cmdHandler.GetCommandHistory).Methods("GET")
	api.HandleFunc("/command/info", cmdHandler.GetShellInfo).Methods("GET")

	api.HandleFunc("/system/processes", sysHandler.GetProcesses).Methods("GET")
	api.HandleFunc("/system/kill", sysHandler.KillProcess).Methods("POST", "OPTIONS")
	api.HandleFunc("/system/cpu", sysHandler.GetCPUUsage).Methods("GET")
	api.HandleFunc("/system/memory", sysHandler.GetMemoryUsage).Methods("GET")

	r.HandleFunc("/ws", wsHandler.HandleWebSocket)

	frontendDevUrl := os.Getenv("DARDCOR_DEV_URL")
	if frontendDevUrl != "" {
		log.Printf("🚀 Dev Proxy Mode: ON")
		log.Printf("🔗 Proxy Target: %s", frontendDevUrl)
		target, _ := url.Parse(frontendDevUrl)
		proxy := httputil.NewSingleHostReverseProxy(target)
		r.PathPrefix("/").HandlerFunc(proxy.ServeHTTP)
	} else {
		log.Printf("📦 Production Mode: Serving static files from ./dist")
		frontendDist := filepath.Join("dist")
		if _, err := os.Stat(frontendDist); err == nil {
			spa := spaHandler{staticPath: frontendDist, indexPath: "index.html"}
			r.PathPrefix("/").Handler(spa)
		} else {
			r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"status":"running","name":"Dardcor Agent","port":"%s","provider":"%s","model":"%s"}`,
					cfg.Port, cfg.AI.Provider, cfg.AI.Model)
			})
		}
	}

	handler := middleware.CORS(middleware.Logger(r))

	addr := "127.0.0.1:" + cfg.Port
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Dardcor Agent")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Dashboard  → http://%s", addr)
	log.Printf("WebSocket  → ws://%s/ws", addr)
	log.Printf("API        → http://%s/api", addr)
	log.Printf("Provider   → %s | %s", cfg.AI.Provider, cfg.AI.Model)
	log.Printf("Data       → %s", cfg.DataDir)
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fpath := filepath.Join(h.staticPath, r.URL.Path)
	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
