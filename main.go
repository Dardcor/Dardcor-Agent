package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
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

//go:embed dist/*
var distFS embed.FS

var startTime = time.Now()

func main() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// ─── CLI Routing ─────────────────────────────────────────────────────────
	args := os.Args[1:]
	isRunCmd := len(args) > 0 && args[0] == "run"

	if !isRunCmd {
		// Only start background engine if no other instance is on this port
		if !isPortInUse(cfg.Port) {
			go startServer(cfg)
			// Settle time for port binding
			time.Sleep(600 * time.Millisecond)
		}

		// Run Node CLI
		nodeArgs := append([]string{"cli.js"}, args...)
		cmd := exec.Command("node", nodeArgs...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}
		os.Exit(0)
	}

	// ─── Direct Server Mode (dardcor run) ───────────────────────────────────
	startServer(cfg)
}

func startServer(cfg *config.Config) {
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

	costSvc := services.NewCostTrackerService()
	idxSvc := services.NewIndexService()
	lspSvc := services.NewLSPService()
	mcpSvc := services.NewMCPClientService()

	agentSvc := services.NewAgentService(fsSvc, cmdSvc, sysSvc, antigravitySvc, memSvc, skillSvc, orchestratorSvc, egoSvc, lspSvc, mcpSvc, costSvc, idxSvc)

	bgAgentSvc := services.NewBackgroundAgentService(agentSvc)

	fsHandler := handlers.NewFileSystemHandler(fsSvc)
	cmdHandler := handlers.NewCommandHandler(cmdSvc)
	sysHandler := handlers.NewSystemHandler(sysSvc)
	agentHandler := handlers.NewAgentHandler(agentSvc)
	wsHandler := handlers.NewWebSocketHandler(agentSvc, cmdSvc)
	antigravityHandler := handlers.NewAntigravityHandler(antigravitySvc)
	modelHandler := handlers.NewModelHandler()
	bgHandler := handlers.NewBackgroundHandler(bgAgentSvc)
	statsHandler := handlers.NewStatsHandler(costSvc, idxSvc, lspSvc, mcpSvc)

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
			"success": true, "data": skills, "count": len(skills),
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

	api.HandleFunc("/background/submit", bgHandler.Submit).Methods("POST", "OPTIONS")
	api.HandleFunc("/background/task", bgHandler.GetTask).Methods("GET")
	api.HandleFunc("/background/tasks", bgHandler.ListTasks).Methods("GET")
	api.HandleFunc("/background/cancel", bgHandler.CancelTask).Methods("POST", "OPTIONS")
	api.HandleFunc("/background/purge", bgHandler.PurgeCompleted).Methods("POST", "OPTIONS")

	api.HandleFunc("/stats", statsHandler.GetStats).Methods("GET")
	api.HandleFunc("/stats/reset", statsHandler.ResetStats).Methods("POST", "OPTIONS")

	api.HandleFunc("/index/status", statsHandler.GetIndexStatus).Methods("GET")
	api.HandleFunc("/index/build", statsHandler.BuildIndex).Methods("POST", "OPTIONS")
	api.HandleFunc("/index/search", statsHandler.SearchIndex).Methods("GET")

	api.HandleFunc("/lsp/status", statsHandler.GetLSPStatus).Methods("GET")

	api.HandleFunc("/mcp/servers", statsHandler.GetMCPServers).Methods("GET")
	api.HandleFunc("/mcp/servers", statsHandler.AddMCPServer).Methods("POST", "OPTIONS")
	api.HandleFunc("/mcp/servers", statsHandler.RemoveMCPServer).Methods("DELETE")

	api.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		convs, _ := storage.Store.ListConversations("web")
		cliConvs, _ := storage.Store.ListConversations("cli")
		all := append(convs, cliConvs...)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": all, "count": len(all)})
	}).Methods("GET")

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
		subFS, _ := fs.Sub(distFS, "dist")
		staticHandler := http.FileServer(http.FS(subFS))
		r.PathPrefix("/assets/").Handler(staticHandler)
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			path := req.URL.Path
			cleanPath := strings.TrimPrefix(path, "/")
			if path != "/" && path != "/index.html" {
				if _, err := fs.Stat(subFS, cleanPath); err == nil {
					staticHandler.ServeHTTP(w, req)
					return
				}
			}
			indexData, _ := fs.ReadFile(subFS, "index.html")
			w.Header().Set("Content-Type", "text/html")
			w.Write(indexData)
		})
	}

	addr := "127.0.0.1:" + cfg.Port
	handler := middleware.CORS(middleware.Logger(r))

	log.SetOutput(os.Stderr)
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Dardcor Agent active on Port %s", cfg.Port)
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func isPortInUse(port string) bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
