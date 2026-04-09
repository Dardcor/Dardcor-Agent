package services

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	AntigravityUserAgent = "vscode/1.92.2 (Antigravity/4.1.31)"
	AntigravityFullUA    = "Antigravity/4.1.31 (Windows NT 10.0; Win64; x64) Chrome/132.0.6834.160 Electron/39.2.3"
)

func getClientID() string {
	if v := os.Getenv("ANTIGRAVITY_CLIENT_ID"); v != "" {
		return v
	}
	return ""
}

func getClientSecret() string {
	if v := os.Getenv("ANTIGRAVITY_CLIENT_SECRET"); v != "" {
		return v
	}
	return ""
}

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
	
	// Ensure directory exists
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
			if err := json.Unmarshal(data, &fileAcc); err != nil {
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

			// Restore Expiry from file so auto-refresh logic works correctly
			if fileAcc.Token.Expiry > 0 {
				acc.Expiry = time.Unix(fileAcc.Token.Expiry, 0)
				// If access token is already expired, clear it to force a refresh on next use
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

			// Deduplicate quotas by name on load (keep highest percentage), restore Key
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
						Key:        m.Key, // restore raw API key
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
}

func (s *AntigravityService) saveAccountFile(acc models.AntigravityAccount) error {
	dir := filepath.Dir(s.dbPath)
	filePath := filepath.Join(dir, acc.ID+".json")

	// Create structure to match exactly
	var fileAcc models.AntigravityFileAccount
	
	// Read existing if any to preserve other fields like device profile
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

func (s *AntigravityService) RefreshToken(email string) (*models.AntigravityAccount, error) {
	s.mu.Lock()
	var account *models.AntigravityAccount
	for i := range s.accounts {
		if s.accounts[i].Email == email {
			account = &s.accounts[i]
			break
		}
	}
	s.mu.Unlock()

	if account == nil {
		return nil, fmt.Errorf("account not found: %s", email)
	}

	// 1. Get Access Token from Google
	data := url.Values{}
	data.Set("client_id", getClientID())
	data.Set("client_secret", getClientSecret())
	data.Set("refresh_token", account.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, _ := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", AntigravityUserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp models.GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	s.mu.Lock()
	account.AccessToken = tokenResp.AccessToken
	account.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	s.mu.Unlock()

	// 2. Load Project ID via loadCodeAssist
	if err := s.fetchProjectID(account); err != nil {
		return nil, err
	}

	// 3. Fetch Quotas
	if err := s.FetchQuotas(account); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.saveAccounts()
	s.mu.Unlock()

	return account, nil
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

    // Logic replicated from Antigravity Manager Tier Extraction:
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

		// Success! Map to internal models with backend deduplication by displayName
		acc.Quotas = []models.ModelQuota{}
		byDisplayName := make(map[string]models.ModelQuota)
		for name, mInfo := range modelsResp.Models {
			// Only include Antigravity-targeted models
			nameLower := strings.ToLower(name)
			isTarget := strings.HasPrefix(nameLower, "gpt") ||
				strings.HasPrefix(nameLower, "image") ||
				strings.HasPrefix(nameLower, "gemini") ||
				strings.HasPrefix(nameLower, "claude")
			if !isTarget {
				continue
			}

			percentage := int(mInfo.QuotaInfo.RemainingFraction * 100)
			if percentage < 0 { percentage = 0 }
			if percentage > 100 { percentage = 100 }

			color := "#7c3aed" // Default purple
			if percentage < 20 {
				color = "#ef4444" // Red
			} else if percentage < 60 {
				color = "#f59e0b" // Orange
			} else {
				color = "#10b981" // Green
			}

			displayName := mInfo.DisplayName
			if displayName == "" {
				displayName = name // Fallback to API key if no display name
			}

			duration := ""
			if mInfo.QuotaInfo.ResetTime != "" {
				duration = mInfo.QuotaInfo.ResetTime
			}

			// Deduplicate: keep entry with highest remaining percentage
			if existing, ok := byDisplayName[displayName]; !ok || percentage > existing.Percentage {
				byDisplayName[displayName] = models.ModelQuota{
					Name:       displayName,
					Key:        name, // store raw API key for use in generateContent
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

	newID := fmt.Sprintf("%d", time.Now().UnixNano()) // using simple timestamp string for new inserts
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
			s.accounts[i].IsActive = !s.accounts[i].IsActive // Toggle
			found = true
		} else {
			s.accounts[i].IsActive = false // Disable others
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
			// Return a copy to prevent mutation
			cpy := acc
			return &cpy, nil
		}
	}
	return nil, fmt.Errorf("no active agent configured. please select an active agent in the Antigravity Model dashboard")
}

// LoadConfig reads persisted Antigravity chat configuration from config.json, or creates it.
func (s *AntigravityService) LoadConfig() models.AntigravityConfig {
	configPath := filepath.Join(filepath.Dir(s.dbPath), "config.json")
	cfg := models.AntigravityConfig{
		Temperature: 0.7,
		MaxTokens:   8192,
		SelectedModel: "",
	}
	
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &cfg)
	} else if os.IsNotExist(err) {
		// Auto-create to populate real-time
		s.SaveConfig(cfg)
	}
	return cfg
}

// SaveConfig persists Antigravity chat configuration to config.json.
func (s *AntigravityService) SaveConfig(cfg models.AntigravityConfig) error {
	configPath := filepath.Join(filepath.Dir(s.dbPath), "config.json")
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 8192
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// FetchProjectAndQuotas fetches project ID and quotas for an account that is missing them.
// It updates the passed account struct, syncs to in-memory accounts, and persists to disk.
func (s *AntigravityService) FetchProjectAndQuotas(acc *models.AntigravityAccount) error {
	if err := s.fetchProjectID(acc); err != nil {
		return err
	}
	if err := s.FetchQuotas(acc); err != nil {
		return err
	}
	// Sync updated fields back into s.accounts
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
	data := url.Values{}
	data.Set("client_id", getClientID())
	data.Set("client_secret", getClientSecret())
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

	// Fetch Email Address matching userinfo scope
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

	// Check for duplicate email BEFORE adding
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

	// Fetch project ID and quotas on the local copy
	s.fetchProjectID(&newAcc)
	s.FetchQuotas(&newAcc)

	// Sync the updated local copy back into s.accounts
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
