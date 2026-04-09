package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dardcor-agent/models"
	"dardcor-agent/services"
)

type AntigravityHandler struct {
	service *services.AntigravityService
}

func NewAntigravityHandler(svc *services.AntigravityService) *AntigravityHandler {
	return &AntigravityHandler{service: svc}
}

func (h *AntigravityHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := h.service.GetAccounts()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    accounts,
	})
}

func (h *AntigravityHandler) RefreshAccount(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "email is required",
		})
		return
	}

	account, err := h.service.RefreshToken(email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    account,
	})
}

func (h *AntigravityHandler) AddAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email        string `json:"email"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if err := h.service.AddAccount(req.Email, req.RefreshToken); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Account added successfully",
	})
}

func (h *AntigravityHandler) RemoveAccount(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "email is required",
		})
		return
	}

	if err := h.service.RemoveAccount(email); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Account removed successfully",
	})
}

func (h *AntigravityHandler) ToggleActiveAccount(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{ Success: false, Error: "email required" })
		return
	}
	if err := h.service.SetActiveAccount(email); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{ Success: false, Error: err.Error() })
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{ Success: true, Message: "Agent activated updated" })
}

func (h *AntigravityHandler) OAuthStart(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent&include_granted_scopes=true", 
		services.AntigravityClientID(), 
		"http%3A%2F%2F127.0.0.1%3A25000%2Fapi%2Fantigravity%2Foauth%2Fcallback", 
		"https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fcloud-platform+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.profile+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fcclog+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fexperimentsandconfigs",
	)
	
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AntigravityHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		redirectURI := "http://127.0.0.1:25000/api/antigravity/oauth/callback"
		err := h.service.ExchangeCode(code, redirectURI)
		if err != nil {
			// If the account already exists, redirect back — user just needs the dashboard
			if len(err.Error()) > 17 && err.Error()[:17] == "account already r" {
				http.Redirect(w, r, "/?error=account_exists", http.StatusTemporaryRedirect)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;background:#0d0920;color:#f8fafc;display:flex;align-items:center;justify-content:center;height:100vh;margin:0">
<div style="text-align:center;max-width:480px;padding:32px;background:#1a0a2e;border:1px solid rgba(124,58,237,.3);border-radius:16px">
<h2 style="color:#ef4444">Authentication Failed</h2>
<p style="color:#94a3b8">%v</p>
<a href="/" style="display:inline-block;margin-top:16px;padding:10px 24px;background:#7c3aed;color:#fff;text-decoration:none;border-radius:8px">← Back to Dashboard</a>
</div></body></html>`, err)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *AntigravityHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.service.LoadConfig()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    cfg,
	})
}

func (h *AntigravityHandler) SaveConfig(w http.ResponseWriter, r *http.Request) {
	var cfg models.AntigravityConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid request body"})
		return
	}
	if err := h.service.SaveConfig(cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Config saved"})
}
