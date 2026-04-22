package services

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RateLimitBucket struct {
	Limit        int
	Remaining    int
	ResetSeconds float64
	CapturedAt   time.Time
}

func (b RateLimitBucket) Used() int {
	return b.Limit - b.Remaining
}

func (b RateLimitBucket) UsagePct() float64 {
	if b.Limit <= 0 {
		return 0
	}
	used := float64(b.Limit - b.Remaining)
	if used < 0 {
		used = 0
	}
	return used / float64(b.Limit)
}

func (b RateLimitBucket) RemainingSecondsNow() float64 {
	if b.CapturedAt.IsZero() {
		return b.ResetSeconds
	}
	elapsed := time.Since(b.CapturedAt).Seconds()
	remaining := b.ResetSeconds - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

type RateLimitState struct {
	RequestsMin  RateLimitBucket
	RequestsHour RateLimitBucket
	TokensMin    RateLimitBucket
	TokensHour   RateLimitBucket
	CapturedAt   time.Time
	Provider     string
}

func (s *RateLimitState) HasData() bool {
	if s == nil {
		return false
	}
	return s.RequestsMin.Limit > 0 ||
		s.RequestsHour.Limit > 0 ||
		s.TokensMin.Limit > 0 ||
		s.TokensHour.Limit > 0
}

func (s *RateLimitState) AgeSeconds() float64 {
	if s == nil || s.CapturedAt.IsZero() {
		return 0
	}
	return time.Since(s.CapturedAt).Seconds()
}

type RateLimiterService struct {
	mu    sync.RWMutex
	state *RateLimitState
}

func NewRateLimiterService() *RateLimiterService {
	return &RateLimiterService{}
}

func (r *RateLimiterService) UpdateFromHeaders(headers map[string]string, provider string) {
	state := ParseRateLimitHeaders(headers, provider)
	if state == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = state
}

func (r *RateLimiterService) GetState() *RateLimitState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func ParseRateLimitHeaders(headers map[string]string, provider string) *RateLimitState {
	now := time.Now()
	state := &RateLimitState{
		CapturedAt: now,
		Provider:   provider,
	}

	norm := make(map[string]string, len(headers))
	for k, v := range headers {
		norm[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}

	parseInt := func(key string) int {
		v, ok := norm[key]
		if !ok {
			return 0
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return n
	}

	parseReset := func(key string) float64 {
		v, ok := norm[key]
		if !ok {
			return 0
		}
		return parseDurationToSeconds(v)
	}

	state.RequestsMin.Limit = parseInt("x-ratelimit-limit-requests")
	state.RequestsMin.Remaining = parseInt("x-ratelimit-remaining-requests")
	state.RequestsMin.ResetSeconds = parseReset("x-ratelimit-reset-requests")
	state.RequestsMin.CapturedAt = now

	state.RequestsHour.Limit = parseInt("x-ratelimit-limit-requests-1h")
	state.RequestsHour.Remaining = parseInt("x-ratelimit-remaining-requests-1h")
	state.RequestsHour.ResetSeconds = parseReset("x-ratelimit-reset-requests-1h")
	state.RequestsHour.CapturedAt = now

	state.TokensMin.Limit = parseInt("x-ratelimit-limit-tokens")
	state.TokensMin.Remaining = parseInt("x-ratelimit-remaining-tokens")
	state.TokensMin.ResetSeconds = parseReset("x-ratelimit-reset-tokens")
	state.TokensMin.CapturedAt = now

	state.TokensHour.Limit = parseInt("x-ratelimit-limit-tokens-1h")
	state.TokensHour.Remaining = parseInt("x-ratelimit-remaining-tokens-1h")
	state.TokensHour.ResetSeconds = parseReset("x-ratelimit-reset-tokens-1h")
	state.TokensHour.CapturedAt = now

	if !state.HasData() {
		return nil
	}
	return state
}

func parseDurationToSeconds(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		return n
	}
	total := 0.0
	rem := s
	if idx := strings.Index(rem, "h"); idx >= 0 {
		h, err := strconv.ParseFloat(rem[:idx], 64)
		if err == nil {
			total += h * 3600
		}
		rem = rem[idx+1:]
	}
	if idx := strings.Index(rem, "m"); idx >= 0 {
		m, err := strconv.ParseFloat(rem[:idx], 64)
		if err == nil {
			total += m * 60
		}
		rem = rem[idx+1:]
	}
	if idx := strings.Index(rem, "s"); idx >= 0 {
		sec, err := strconv.ParseFloat(rem[:idx], 64)
		if err == nil {
			total += sec
		}
	}
	return total
}

func fmtCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 10_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return strconv.Itoa(n)
	}
}

func fmtSeconds(seconds float64) string {
	if seconds <= 0 {
		return "0s"
	}
	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0 && s > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}

func bar(pct float64, width int) string {
	if width <= 0 {
		width = 20
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(math.Round(pct * float64(width)))
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func bucketLine(label string, b RateLimitBucket) string {
	if b.Limit <= 0 {
		return ""
	}
	pct := b.UsagePct()
	resetIn := b.RemainingSecondsNow()
	warning := ""
	if pct >= 0.80 {
		warning = " ⚠"
	}
	return fmt.Sprintf("  %-18s %s %5.1f%%  used:%s/%s  reset:%s%s",
		label+":",
		bar(pct, 20),
		pct*100,
		fmtCount(b.Used()),
		fmtCount(b.Limit),
		fmtSeconds(resetIn),
		warning,
	)
}

func FormatRateLimitDisplay(state *RateLimitState) string {
	if state == nil || !state.HasData() {
		return ""
	}

	var sb strings.Builder
	provider := state.Provider
	if provider == "" {
		provider = "unknown"
	}
	age := state.AgeSeconds()
	sb.WriteString(fmt.Sprintf("┌─ Rate Limits · %s  (captured %.0fs ago)\n", provider, age))

	rows := []struct {
		label  string
		bucket RateLimitBucket
	}{
		{"Req/min", state.RequestsMin},
		{"Req/hour", state.RequestsHour},
		{"Tok/min", state.TokensMin},
		{"Tok/hour", state.TokensHour},
	}

	anyWarning := false
	for _, row := range rows {
		line := bucketLine(row.label, row.bucket)
		if line == "" {
			continue
		}
		sb.WriteString(line)
		sb.WriteByte('\n')
		if row.bucket.UsagePct() >= 0.80 {
			anyWarning = true
		}
	}

	if anyWarning {
		sb.WriteString("  ⚠  One or more buckets are ≥80% consumed — approaching rate limit.\n")
	}
	sb.WriteString("└────────────────────────────────────────────────────────────\n")
	return sb.String()
}

func FormatRateLimitCompact(state *RateLimitState) string {
	if state == nil || !state.HasData() {
		return ""
	}

	parts := make([]string, 0, 4)

	addBucket := func(label string, b RateLimitBucket) {
		if b.Limit <= 0 {
			return
		}
		parts = append(parts, fmt.Sprintf("%s %s/%s (%.1f%%)",
			label,
			fmtCount(b.Used()),
			fmtCount(b.Limit),
			b.UsagePct()*100,
		))
	}

	if state.RequestsHour.Limit > 0 {
		addBucket("Req", state.RequestsHour)
	} else if state.RequestsMin.Limit > 0 {
		addBucket("Req/m", state.RequestsMin)
	}

	if state.TokensHour.Limit > 0 {
		addBucket("Tok", state.TokensHour)
	} else if state.TokensMin.Limit > 0 {
		addBucket("Tok/m", state.TokensMin)
	}

	if len(parts) == 0 {
		return ""
	}

	provider := state.Provider
	if provider == "" {
		provider = "?"
	}

	summary := fmt.Sprintf("[%s] %s", provider, strings.Join(parts, "  "))

	for _, b := range []RateLimitBucket{state.RequestsMin, state.RequestsHour, state.TokensMin, state.TokensHour} {
		if b.UsagePct() >= 0.80 {
			summary += " ⚠"
			break
		}
	}
	return summary
}
