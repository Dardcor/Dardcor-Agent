package models

import "time"

type AntigravityAccount struct {
	ID           string       `json:"id"`
	Email        string       `json:"email"`
	RefreshToken string       `json:"refresh_token"`
	AccessToken  string       `json:"access_token,omitempty"`
	Expiry       time.Time    `json:"expiry,omitempty"`
	ProjectID    string       `json:"project_id,omitempty"`
	Type         string       `json:"type"`
	Status       string       `json:"status"`
	IsActive     bool         `json:"is_active"`
	Quotas       []ModelQuota `json:"quotas,omitempty"`
	LastUsed     time.Time    `json:"last_used"`
}

type AntigravityFileAccount struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expiry       int64  `json:"expiry_timestamp"`
		ProjectID    string `json:"project_id"`
	} `json:"token"`
	Quota struct {
		Models []struct {
			Name       string `json:"name"`
			Key        string `json:"key,omitempty"`
			Percentage int    `json:"percentage"`
			ResetTime  string `json:"reset_time"`
		} `json:"models"`
		SubscriptionTier string `json:"subscription_tier"`
		IsForbidden      bool   `json:"is_forbidden"`
	} `json:"quota"`
	IsActive bool  `json:"is_active"`
	LastUsed int64 `json:"last_used"`
}

type ModelQuota struct {
	Name       string `json:"name"`
	Key        string `json:"key,omitempty"`
	Percentage int    `json:"percentage"`
	Available  bool   `json:"available"`
	Duration   string `json:"duration,omitempty"`
	Color      string `json:"color,omitempty"`
}


type AntigravityAuth struct {
	GoogleClientID     string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type AntigravityLoadAssistResponse struct {
	ProjectID   string `json:"cloudaicompanionProject"`
	CurrentTier struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"currentTier"`
	PaidTier struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"paidTier"`
	AllowedTiers []struct {
		IsDefault bool   `json:"is_default"`
		ID        string `json:"id"`
		Name      string `json:"name"`
	} `json:"allowedTiers"`
	IneligibleTiers []struct {
		ReasonCode string `json:"reasonCode"`
	} `json:"ineligibleTiers"`
}

type ModelInfo struct {
	QuotaInfo struct {
		RemainingFraction float64 `json:"remainingFraction"`
		ResetTime         string  `json:"resetTime"`
	} `json:"quotaInfo"`
	DisplayName string `json:"displayName"`
}

type AntigravityAvailableModelsResponse struct {
	Models map[string]ModelInfo `json:"models"`
}
