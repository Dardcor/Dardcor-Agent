package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	memSvc := services.NewMemoryService(cfg.DataDir)
	fsSvc := services.NewFileSystemService()
	cmdSvc := services.NewCommandService()
	sysSvc := services.NewSystemService()
	antigravitySvc := services.NewAntigravityService()
	egoSvc := services.NewEgoService()
	dreamSvc := services.NewDreamService(fsSvc, egoSvc)
	dreamSvc.StartDreaming()
	skillSvc := services.NewSkillService()
	orchestratorSvc := services.NewOrchestratorService()
	reflectSvc := services.NewReflectionService(egoSvc, orchestratorSvc)
	browserSvc := services.NewBrowserService()
	visionSvc := services.NewVisionService()
	autoSvc := services.NewAutomationService()
	agentSvc := services.NewAgentService(fsSvc, cmdSvc, sysSvc, antigravitySvc, memSvc, skillSvc, orchestratorSvc, egoSvc, reflectSvc, browserSvc, visionSvc, autoSvc)

	fsHandler := handlers.NewFileSystemHandler(fsSvc)
	cmdHandler := handlers.NewCommandHandler(cmdSvc)
	sysHandler := handlers.NewSystemHandler(sysSvc)
	agentHandler := handlers.NewAgentHandler(agentSvc)
	wsHandler := handlers.NewWebSocketHandler(agentSvc, cmdSvc)

	antigravityHandler := handlers.NewAntigravityHandler(antigravitySvc)
	modelHandler := handlers.NewModelHandler()

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
		skills := skillSvc.GetSkills()
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
	api.HandleFunc("/files/workspace/default", fsHandler.GetDefaultWorkspace).Methods("GET")
	api.HandleFunc("/files/move", fsHandler.MoveFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/files/copy", fsHandler.CopyFile).Methods("POST", "OPTIONS")

	api.HandleFunc("/command", cmdHandler.ExecuteCommand).Methods("POST", "OPTIONS")
	api.HandleFunc("/command/history", cmdHandler.GetCommandHistory).Methods("GET")
	api.HandleFunc("/command/info", cmdHandler.GetShellInfo).Methods("GET")

	api.HandleFunc("/system/processes", sysHandler.GetProcesses).Methods("GET")
	api.HandleFunc("/system/kill", sysHandler.KillProcess).Methods("POST", "OPTIONS")
	api.HandleFunc("/system/cpu", sysHandler.GetCPUUsage).Methods("GET")
	api.HandleFunc("/system/memory", sysHandler.GetMemoryUsage).Methods("GET")

	api.HandleFunc("/antigravity/accounts", antigravityHandler.GetAccounts).Methods("GET")
	api.HandleFunc("/antigravity/accounts", antigravityHandler.AddAccount).Methods("POST", "OPTIONS")
	api.HandleFunc("/antigravity/accounts", antigravityHandler.RemoveAccount).Methods("DELETE")
	api.HandleFunc("/antigravity/refresh", antigravityHandler.RefreshAccount).Methods("GET")
	api.HandleFunc("/antigravity/oauth/start", antigravityHandler.OAuthStart).Methods("GET")
	api.HandleFunc("/antigravity/oauth/callback", antigravityHandler.OAuthCallback).Methods("GET")
	api.HandleFunc("/antigravity/active", antigravityHandler.ToggleActiveAccount).Methods("POST", "OPTIONS")
	api.HandleFunc("/antigravity/config", antigravityHandler.GetAccounts).Methods("GET")
	api.HandleFunc("/antigravity/config", antigravityHandler.SaveConfig).Methods("POST", "OPTIONS")

	api.HandleFunc("/model/active", modelHandler.GetModelConfig).Methods("GET")
	api.HandleFunc("/model/active", modelHandler.SaveModelConfig).Methods("POST", "OPTIONS")

	api.HandleFunc("/tools/config", modelHandler.GetToolsConfig).Methods("GET")
	api.HandleFunc("/tools/config", modelHandler.SaveToolsConfig).Methods("POST", "OPTIONS")

	api.HandleFunc("/skills/config", modelHandler.GetSkillsConfig).Methods("GET")
	api.HandleFunc("/skills/config", modelHandler.SaveSkillsConfig).Methods("POST", "OPTIONS")

	api.HandleFunc("/workspace/config", modelHandler.GetWorkspaceConfig).Methods("GET")
	api.HandleFunc("/workspace/config", modelHandler.SaveWorkspaceConfig).Methods("POST", "OPTIONS")
	egoHandler := handlers.NewEgoHandler(egoSvc, dreamSvc)
	api.HandleFunc("/ego/state", egoHandler.GetState).Methods("GET")
	api.HandleFunc("/ego/dreams", egoHandler.GetDreams).Methods("GET")

	storagePath := filepath.Join(cfg.DataDir, "storage")
	r.PathPrefix("/storage/").Handler(http.StripPrefix("/storage/", http.FileServer(http.Dir(storagePath))))

	r.HandleFunc("/ws", wsHandler.HandleWebSocket)

	isDev := os.Getenv("DARDCOR_DEV") == "true"

	if isDev {
		target, _ := url.Parse("http://127.0.0.1:5173")
		proxy := httputil.NewSingleHostReverseProxy(target)
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/api") || req.URL.Path == "/ws" {
				return
			}
			proxy.ServeHTTP(w, req)
		})
	} else {
		r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("dist/assets"))))
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			path := req.URL.Path
			if path != "/" && path != "/index.html" {
				if _, err := os.Stat("dist" + path); err == nil {
					http.ServeFile(w, req, "dist"+path)
					return
				}
			}
			http.ServeFile(w, req, "dist/index.html")
		})
	}

	handler := middleware.CORS(middleware.Logger(r))

	isDev = os.Getenv("DARDCOR_DEV") == "true"

	addr := "127.0.0.1:" + cfg.Port

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
