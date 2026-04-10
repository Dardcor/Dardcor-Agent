package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"
)

const (
	AntigravityUserAgent = "vscode (Antigravity)"
	AntigravityFullUA    = "Antigravity (Windows NT; Win64; x64) Chrome Electron"
)

// Standalone credential functions removed. Use AntigravityService.LoadConfig() instead.

type GoogleUserInfo struct {
	Email string `json:"email"`
}

type AntigravityService struct {
	accounts []models.AntigravityAccount
	mu       sync.Mutex
	dbPath   string
	client   *http.Client
}

func NewAntigravityService() *AntigravityService {
	baseDir := "database"
	if config.AppConfig != nil && config.AppConfig.DataDir != "" {
		baseDir = config.AppConfig.DataDir
	}
	dbPath := filepath.Join(baseDir, "model", "antigravity", "accounts.json")

	os.MkdirAll(filepath.Dir(dbPath), 0755)

	svc := &AntigravityService{
		dbPath: dbPath,
		client: &http.Client{Timeout: 30 * time.Second},
	}
	svc.loadAccounts()
	return svc
}

func (s *AntigravityService) loadAccounts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accounts = []models.AntigravityAccount{}
	dir := filepath.Dir(s.dbPath)

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" && file.Name() != "accounts.json" {
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				continue
			}

			var fileAcc models.AntigravityFileAccount
			if err := json.Unmarshal(data, &fileAcc); err != nil || fileAcc.Email == "" || fileAcc.Token.RefreshToken == "" {
				continue
			}

			acc := models.AntigravityAccount{
				ID:           fileAcc.ID,
				Email:        fileAcc.Email,
				RefreshToken: fileAcc.Token.RefreshToken,
				AccessToken:  fileAcc.Token.AccessToken,
				ProjectID:    fileAcc.Token.ProjectID,
				Type:         fileAcc.Quota.SubscriptionTier,
				IsActive:     fileAcc.IsActive,
			}

			if fileAcc.Token.Expiry > 0 {
				acc.Expiry = time.Unix(fileAcc.Token.Expiry, 0)

				if time.Now().After(acc.Expiry) {
					acc.AccessToken = ""
				}
			}

			if fileAcc.Quota.IsForbidden {
				acc.Status = "FORBIDDEN"
			} else if fileAcc.Token.AccessToken != "" {
				acc.Status = "CURRENT"
			}

			if acc.Type == "" {
				acc.Type = "FREE"
			}

			quotaMap := make(map[string]models.ModelQuota)
			for _, m := range fileAcc.Quota.Models {
				color := "#7c3aed"
				if m.Percentage < 20 {
					color = "#ef4444"
				} else if m.Percentage < 60 {
					color = "#f59e0b"
				} else {
					color = "#10b981"
				}
				if existing, ok := quotaMap[m.Name]; !ok || m.Percentage > existing.Percentage {
					quotaMap[m.Name] = models.ModelQuota{
						Name:       m.Name,
						Key:        m.Key,
						Percentage: m.Percentage,
						Available:  m.Percentage > 0,
						Color:      color,
						Duration:   m.ResetTime,
					}
				}
			}
			for _, q := range quotaMap {
				acc.Quotas = append(acc.Quotas, q)
			}
			acc.LastUsed = time.Unix(fileAcc.LastUsed, 0)
			s.accounts = append(s.accounts, acc)
		}
	}
	fmt.Printf("[AntigravityService] Loaded %d accounts from %s\n", len(s.accounts), dir)
}

func (s *AntigravityService) saveAccountFile(acc models.AntigravityAccount) error {
	dir := filepath.Dir(s.dbPath)
	filePath := filepath.Join(dir, acc.ID+".json")

	var fileAcc models.AntigravityFileAccount

	data, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(data, &fileAcc)
	}

	fileAcc.ID = acc.ID
	fileAcc.Email = acc.Email
	if fileAcc.Name == "" {
		fileAcc.Name = acc.Email
	}

	fileAcc.Token.AccessToken = acc.AccessToken
	fileAcc.Token.RefreshToken = acc.RefreshToken
	fileAcc.Token.ProjectID = acc.ProjectID
	fileAcc.Token.Expiry = acc.Expiry.Unix()

	fileAcc.Quota.SubscriptionTier = acc.Type
	fileAcc.Quota.IsForbidden = (acc.Status == "FORBIDDEN")

	fileAcc.IsActive = acc.IsActive
	fileAcc.LastUsed = acc.LastUsed.Unix()
	fileAcc.Quota.Models = nil
	for _, mq := range acc.Quotas {
		fileAcc.Quota.Models = append(fileAcc.Quota.Models, struct {
			Name       string `json:"name"`
			Key        string `json:"key,omitempty"`
			Percentage int    `json:"percentage"`
			ResetTime  string `json:"reset_time"`
		}{
			Name:       mq.Name,
			Key:        mq.Key,
			Percentage: mq.Percentage,
			ResetTime:  mq.Duration,
		})
	}
	fileAcc.LastUsed = acc.LastUsed.Unix()

	outData, _ := json.MarshalIndent(fileAcc, "", "  ")
	return os.WriteFile(filePath, outData, 0644)
}

func (s *AntigravityService) saveAccounts() error {
	for _, acc := range s.accounts {
		s.saveAccountFile(acc)
	}
	return nil
}

func (s *AntigravityService) fetchProjectID(acc *models.AntigravityAccount) error {
	payload := map[string]interface{}{
		"metadata": map[string]string{
			"ideType": "ANTIGRAVITY",
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://daily-cloudcode-pa.sandbox.googleapis.com/v1internal:loadCodeAssist", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+acc.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", AntigravityUserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to load code assist: %d", resp.StatusCode)
	}

	var assistResp models.AntigravityLoadAssistResponse
	if err := json.NewDecoder(resp.Body).Decode(&assistResp); err != nil {
		return err
	}

	acc.ProjectID = assistResp.ProjectID

	subscriptionTier := ""
	isIneligible := len(assistResp.IneligibleTiers) > 0

	if assistResp.PaidTier.Name != "" {
		subscriptionTier = assistResp.PaidTier.Name
	} else if assistResp.PaidTier.ID != "" {
		subscriptionTier = assistResp.PaidTier.ID
	}

	if subscriptionTier == "" {
		if !isIneligible {
			if assistResp.CurrentTier.Name != "" {
				subscriptionTier = assistResp.CurrentTier.Name
			} else if assistResp.CurrentTier.ID != "" {
				subscriptionTier = assistResp.CurrentTier.ID
			}
		} else {
			for _, t := range assistResp.AllowedTiers {
				if t.IsDefault {
					if t.Name != "" {
						subscriptionTier = t.Name + " (Restricted)"
					} else {
						subscriptionTier = t.ID + " (Restricted)"
					}
					break
				}
			}
		}
	}

	if subscriptionTier != "" {
		acc.Type = subscriptionTier
	} else if acc.Type == "" {
		acc.Type = "FREE"
	}

	return nil
}

func (s *AntigravityService) FetchQuotas(acc *models.AntigravityAccount) error {
	if acc.ProjectID == "" {
		return fmt.Errorf("project ID is missing")
	}

	payload := map[string]string{
		"project": acc.ProjectID,
	}
	body, _ := json.Marshal(payload)

	// Quota API endpoints (fallback order: Sandbox → Daily → Prod)
	endpoints := []string{
		"https://daily-cloudcode-pa.sandbox.googleapis.com/v1internal:fetchAvailableModels",
		"https://daily-cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels",
		"https://cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels",
	}

	var lastErr error
	for _, urlStr := range endpoints {
		req, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+acc.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", AntigravityUserAgent)

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusForbidden {
			acc.Status = "FORBIDDEN"
			return fmt.Errorf("account forbidden (403 HTTP)")
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API error %d from %s", resp.StatusCode, urlStr)
			continue
		}

		acc.Status = "CURRENT"

		var modelsResp models.AntigravityAvailableModelsResponse
		if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
			lastErr = err
			continue
		}

		acc.Quotas = []models.ModelQuota{}
		byDisplayName := make(map[string]models.ModelQuota)
		for name, mInfo := range modelsResp.Models {

			nameLower := strings.ToLower(name)
			isTarget := strings.HasPrefix(nameLower, "gpt") ||
				strings.HasPrefix(nameLower, "image") ||
				strings.HasPrefix(nameLower, "imagen") ||
				strings.HasPrefix(nameLower, "gemini") ||
				strings.HasPrefix(nameLower, "claude")
			if !isTarget {
				continue
			}

			percentage := int(mInfo.QuotaInfo.RemainingFraction * 100)
			if percentage < 0 {
				percentage = 0
			}
			if percentage > 100 {
				percentage = 100
			}

			color := "#7c3aed"
			if percentage < 20 {
				color = "#ef4444"
			} else if percentage < 60 {
				color = "#f59e0b"
			} else {
				color = "#10b981"
			}

			displayName := mInfo.DisplayName
			if displayName == "" {
				displayName = name
			}

			duration := ""
			if mInfo.QuotaInfo.ResetTime != "" {
				duration = mInfo.QuotaInfo.ResetTime
			}

			if existing, ok := byDisplayName[displayName]; !ok || percentage > existing.Percentage {
				byDisplayName[displayName] = models.ModelQuota{
					Name:       displayName,
					Key:        name,
					Percentage: percentage,
					Available:  percentage > 0,
					Color:      color,
					Duration:   duration,
				}
			}
		}
		for _, q := range byDisplayName {
			acc.Quotas = append(acc.Quotas, q)
		}
		acc.LastUsed = time.Now()
		return nil
	}

	return fmt.Errorf("all quota endpoints failed: %v", lastErr)
}

func (s *AntigravityService) GetAccounts() []models.AntigravityAccount {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accounts
}

func (s *AntigravityService) AddAccount(email, refreshToken string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, acc := range s.accounts {
		if acc.Email == email {
			return fmt.Errorf("account already exists")
		}
	}

	newID := fmt.Sprintf("%d", time.Now().UnixNano())
	newAcc := models.AntigravityAccount{
		ID:           newID,
		Email:        email,
		RefreshToken: refreshToken,
		Type:         "FREE",
		LastUsed:     time.Now(),
	}
	s.accounts = append(s.accounts, newAcc)
	return s.saveAccountFile(newAcc)
}

func (s *AntigravityService) RemoveAccount(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, acc := range s.accounts {
		if acc.Email == email {
			filePath := filepath.Join(filepath.Dir(s.dbPath), acc.ID+".json")
			os.Remove(filePath)

			s.accounts = append(s.accounts[:i], s.accounts[i+1:]...)
			return s.saveAccounts()
		}
	}
	return fmt.Errorf("account not found")
}

func (s *AntigravityService) SetActiveAccount(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, acc := range s.accounts {
		if acc.Email == email {
			s.accounts[i].IsActive = !s.accounts[i].IsActive
			found = true
		} else {
			s.accounts[i].IsActive = false
		}
	}

	if !found {
		return fmt.Errorf("account not found")
	}
	return s.saveAccounts()
}

func (s *AntigravityService) GetActiveAccount() (*models.AntigravityAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, acc := range s.accounts {
		if acc.IsActive {

			cpy := acc
			return &cpy, nil
		}
	}
	return nil, fmt.Errorf("no active agent configured. please select an active agent in the Antigravity Model dashboard")
}

func (s *AntigravityService) LoadConfig() models.AntigravityConfig {
	configPath := filepath.Join(filepath.Dir(s.dbPath), "config.json")
	cfg := models.AntigravityConfig{
		Temperature:        0.7,
		MaxTokens:          8192,
		SelectedModel:      "",
		GoogleClientID:     "ENTER_CLIENT_ID_VIA_DASHBOARD",
		GoogleClientSecret: "ENTER_CLIENT_SECRET_VIA_DASHBOARD",
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &cfg)
	} else if os.IsNotExist(err) {
		s.SaveConfig(cfg)
	}
	return cfg
}

func (s *AntigravityService) SaveConfig(cfg models.AntigravityConfig) error {
	configPath := filepath.Join(filepath.Dir(s.dbPath), "config.json")
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 8192
	}
	if cfg.GoogleClientID == "" {
		cfg.GoogleClientID = "ENTER_CLIENT_ID_VIA_DASHBOARD"
	}
	if cfg.GoogleClientSecret == "" {
		cfg.GoogleClientSecret = "ENTER_CLIENT_SECRET_VIA_DASHBOARD"
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

func (s *AntigravityService) FetchProjectAndQuotas(acc *models.AntigravityAccount) error {
	if err := s.fetchProjectID(acc); err != nil {
		return err
	}
	if err := s.FetchQuotas(acc); err != nil {
		return err
	}

	s.mu.Lock()
	for i := range s.accounts {
		if s.accounts[i].Email == acc.Email {
			s.accounts[i].ProjectID = acc.ProjectID
			s.accounts[i].Type = acc.Type
			s.accounts[i].Quotas = acc.Quotas
			s.accounts[i].LastUsed = acc.LastUsed
			break
		}
	}
	s.saveAccountFile(*acc)
	s.mu.Unlock()
	return nil
}

func (s *AntigravityService) ExchangeCode(code string, redirectURI string) error {
	cfg := s.LoadConfig()
	data := url.Values{}
	data.Set("client_id", cfg.GoogleClientID)
	data.Set("client_secret", cfg.GoogleClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", AntigravityFullUA)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token exchange failed: %d", resp.StatusCode)
	}

	var tokenResp models.GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	userReq, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userResp, err := s.client.Do(userReq)
	if err != nil {
		return err
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch userinfo")
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return err
	}

	s.mu.Lock()
	for _, existing := range s.accounts {
		if existing.Email == userInfo.Email {
			s.mu.Unlock()
			return fmt.Errorf("account already registered: %s", userInfo.Email)
		}
	}
	newID := fmt.Sprintf("%d", time.Now().UnixNano())
	newAcc := models.AntigravityAccount{
		ID:           newID,
		Email:        userInfo.Email,
		RefreshToken: tokenResp.RefreshToken,
		AccessToken:  tokenResp.AccessToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Type:         "FREE",
		LastUsed:     time.Now(),
	}
	s.accounts = append(s.accounts, newAcc)
	s.mu.Unlock()

	s.fetchProjectID(&newAcc)
	s.FetchQuotas(&newAcc)

	s.mu.Lock()
	for i := range s.accounts {
		if s.accounts[i].Email == newAcc.Email {
			s.accounts[i] = newAcc
			break
		}
	}
	err = s.saveAccountFile(newAcc)
	s.mu.Unlock()

	return err
}

func (s *AntigravityService) RefreshToken(email string) (*models.AntigravityAccount, error) {
	s.mu.Lock()
	var acc *models.AntigravityAccount
	for i := range s.accounts {
		if s.accounts[i].Email == email {
			acc = &s.accounts[i]
			break
		}
	}
	s.mu.Unlock()

	if acc == nil {
		return nil, fmt.Errorf("account not found")
	}

	cfg := s.LoadConfig()
	data := url.Values{}
	data.Set("client_id", cfg.GoogleClientID)
	data.Set("client_secret", cfg.GoogleClientSecret)
	data.Set("refresh_token", acc.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", AntigravityFullUA)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp models.GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	s.mu.Lock()
	acc.AccessToken = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		acc.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}
	acc.LastUsed = time.Now()
	s.saveAccountFile(*acc)
	s.mu.Unlock()

	return acc, nil
}
