package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"dardcor-agent/config"
	"dardcor-agent/models"
)

type InsightsReport struct {
	Days        int
	Empty       bool
	GeneratedAt time.Time
	Overview    *InsightsOverview
	Models      []ModelBreakdown
	Tools       []ToolBreakdown
	Activity    *ActivityPatterns
	TopSessions []NotableSession
}

type InsightsOverview struct {
	TotalSessions         int
	TotalMessages         int
	TotalToolCalls        int
	TotalInputTokens      int
	TotalOutputTokens     int
	TotalTokens           int
	EstimatedCost         float64
	TotalHours            float64
	AvgSessionDuration    float64
	AvgMessagesPerSession float64
	ActiveDays            int
	DateRangeStart        time.Time
	DateRangeEnd          time.Time
}

type ModelBreakdown struct {
	Model        string
	Sessions     int
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	ToolCalls    int
	Cost         float64
}

type ToolBreakdown struct {
	Tool       string
	Count      int
	Percentage float64
}

type ActivityPatterns struct {
	ByDay       []DayCount
	ByHour      []HourCount
	BusiestDay  *DayCount
	BusiestHour *HourCount
	ActiveDays  int
	MaxStreak   int
}

type DayCount struct {
	Day   string
	Count int
}

type HourCount struct {
	Hour  int
	Count int
}

type NotableSession struct {
	Label     string
	SessionID string
	Value     string
	Date      string
}

type InsightsService struct {
	costTracker *CostTrackerService
}

func NewInsightsService(costTracker *CostTrackerService) *InsightsService {
	return &InsightsService{costTracker: costTracker}
}

type sessionSnapshot struct {
	conv       models.Conversation
	toolCounts map[string]int
	totalTools int
	duration   float64
}

func (s *InsightsService) Generate(days int) *InsightsReport {
	report := &InsightsReport{
		Days:        days,
		GeneratedAt: time.Now(),
	}

	sessions := s.loadSessions(days)

	if len(sessions) == 0 {
		report.Empty = true
		return report
	}

	report.Overview = s.computeOverview(sessions)
	report.Models = s.computeModelBreakdown(sessions)
	report.Tools = s.computeToolBreakdown(sessions)
	report.Activity = s.computeActivityPatterns(sessions)
	report.TopSessions = s.computeTopSessions(sessions)

	return report
}

func (s *InsightsService) loadSessions(days int) []sessionSnapshot {
	dbDir := "database"
	if config.AppConfig != nil && config.AppConfig.DataDir != "" {
		dbDir = config.AppConfig.DataDir
	}

	cutoff := time.Now().AddDate(0, 0, -days)

	entries, err := os.ReadDir(dbDir)
	if err != nil {
		return nil
	}

	var snapshots []sessionSnapshot

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		if e.Name() == "stats.json" {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(dbDir, e.Name()))
		if err != nil {
			continue
		}

		var conv models.Conversation
		if err := json.Unmarshal(raw, &conv); err != nil {
			continue
		}

		ref := conv.UpdatedAt
		if ref.IsZero() {
			ref = conv.CreatedAt
		}
		if ref.Before(cutoff) {
			continue
		}

		snap := sessionSnapshot{
			conv:       conv,
			toolCounts: make(map[string]int),
		}

		for _, msg := range conv.Messages {
			for _, act := range msg.Actions {
				tool := act.Type
				if tool == "" {
					tool = "unknown"
				}
				snap.toolCounts[tool]++
				snap.totalTools++
			}
		}

		if len(conv.Messages) >= 2 {
			first := conv.Messages[0].Timestamp
			last := conv.Messages[len(conv.Messages)-1].Timestamp
			if !first.IsZero() && !last.IsZero() && last.After(first) {
				snap.duration = last.Sub(first).Seconds()
			}
		}

		snapshots = append(snapshots, snap)
	}

	return snapshots
}

func (s *InsightsService) computeOverview(sessions []sessionSnapshot) *InsightsOverview {
	ov := &InsightsOverview{}

	stats := s.costTracker.GetStats()
	ov.TotalInputTokens = stats.TotalInputTokens
	ov.TotalOutputTokens = stats.TotalOutputTokens
	ov.TotalTokens = stats.TotalInputTokens + stats.TotalOutputTokens
	ov.EstimatedCost = stats.TotalCost

	ov.TotalSessions = len(sessions)

	activeDaySet := make(map[string]bool)
	var earliest, latest time.Time

	for _, snap := range sessions {
		ov.TotalMessages += len(snap.conv.Messages)
		ov.TotalToolCalls += snap.totalTools
		ov.TotalHours += snap.duration / 3600.0

		ref := snap.conv.UpdatedAt
		if ref.IsZero() {
			ref = snap.conv.CreatedAt
		}
		if !ref.IsZero() {
			dayKey := ref.Format("2006-01-02")
			activeDaySet[dayKey] = true

			if earliest.IsZero() || ref.Before(earliest) {
				earliest = ref
			}
			if latest.IsZero() || ref.After(latest) {
				latest = ref
			}
		}
	}

	ov.ActiveDays = len(activeDaySet)
	ov.DateRangeStart = earliest
	ov.DateRangeEnd = latest

	if ov.TotalSessions > 0 {
		ov.AvgSessionDuration = ov.TotalHours * 3600.0 / float64(ov.TotalSessions)
		ov.AvgMessagesPerSession = float64(ov.TotalMessages) / float64(ov.TotalSessions)
	}

	return ov
}

func (s *InsightsService) computeModelBreakdown(sessions []sessionSnapshot) []ModelBreakdown {
	stats := s.costTracker.GetStats()

	var result []ModelBreakdown
	for provider, ps := range stats.ByProvider {
		mb := ModelBreakdown{
			Model:        provider,
			Sessions:     ps.Requests,
			InputTokens:  ps.InputTokens,
			OutputTokens: ps.OutputTokens,
			TotalTokens:  ps.InputTokens + ps.OutputTokens,
			Cost:         ps.Cost,
		}
		result = append(result, mb)
	}

	totalTools := 0
	for _, snap := range sessions {
		totalTools += snap.totalTools
	}
	if len(result) == 1 && totalTools > 0 {
		result[0].ToolCalls = totalTools
	} else if len(result) > 1 && totalTools > 0 {
		totalReqs := 0
		for _, mb := range result {
			totalReqs += mb.Sessions
		}
		for i := range result {
			if totalReqs > 0 {
				result[i].ToolCalls = int(math.Round(float64(totalTools) * float64(result[i].Sessions) / float64(totalReqs)))
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalTokens > result[j].TotalTokens
	})

	return result
}

func (s *InsightsService) computeToolBreakdown(sessions []sessionSnapshot) []ToolBreakdown {
	counts := make(map[string]int)
	total := 0

	for _, snap := range sessions {
		for tool, cnt := range snap.toolCounts {
			counts[tool] += cnt
			total += cnt
		}
	}

	var result []ToolBreakdown
	for tool, cnt := range counts {
		pct := 0.0
		if total > 0 {
			pct = float64(cnt) / float64(total) * 100.0
		}
		result = append(result, ToolBreakdown{Tool: tool, Count: cnt, Percentage: pct})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

func (s *InsightsService) computeActivityPatterns(sessions []sessionSnapshot) *ActivityPatterns {
	dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	dayCounts := make([]int, 7)
	hourCounts := make([]int, 24)
	activeDaySet := make(map[string]bool)

	for _, snap := range sessions {
		ref := snap.conv.UpdatedAt
		if ref.IsZero() {
			ref = snap.conv.CreatedAt
		}
		if ref.IsZero() {
			continue
		}

		dayCounts[ref.Weekday()]++
		hourCounts[ref.Hour()]++
		activeDaySet[ref.Format("2006-01-02")] = true
	}

	byDay := make([]DayCount, 7)
	for i := 0; i < 7; i++ {

		wd := (i + 1) % 7
		byDay[i] = DayCount{Day: dayNames[wd], Count: dayCounts[wd]}
	}

	byHour := make([]HourCount, 24)
	for h := 0; h < 24; h++ {
		byHour[h] = HourCount{Hour: h, Count: hourCounts[h]}
	}

	var busiestDay *DayCount
	for i := range byDay {
		if busiestDay == nil || byDay[i].Count > busiestDay.Count {
			cp := byDay[i]
			busiestDay = &cp
		}
	}

	var busiestHour *HourCount
	for i := range byHour {
		if busiestHour == nil || byHour[i].Count > busiestHour.Count {
			cp := byHour[i]
			busiestHour = &cp
		}
	}

	maxStreak := s.calcStreak(activeDaySet)

	return &ActivityPatterns{
		ByDay:       byDay,
		ByHour:      byHour,
		BusiestDay:  busiestDay,
		BusiestHour: busiestHour,
		ActiveDays:  len(activeDaySet),
		MaxStreak:   maxStreak,
	}
}

func (s *InsightsService) calcStreak(daySet map[string]bool) int {
	if len(daySet) == 0 {
		return 0
	}

	days := make([]time.Time, 0, len(daySet))
	for k := range daySet {
		t, err := time.Parse("2006-01-02", k)
		if err == nil {
			days = append(days, t)
		}
	}

	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })

	maxStreak, cur := 1, 1
	for i := 1; i < len(days); i++ {
		diff := days[i].Sub(days[i-1])
		if diff == 24*time.Hour {
			cur++
			if cur > maxStreak {
				maxStreak = cur
			}
		} else {
			cur = 1
		}
	}
	return maxStreak
}

func (s *InsightsService) computeTopSessions(sessions []sessionSnapshot) []NotableSession {
	if len(sessions) == 0 {
		return nil
	}

	type candidate struct {
		snap  sessionSnapshot
		score float64
	}

	pick := func(label string, score func(sessionSnapshot) float64, format func(sessionSnapshot) string) *NotableSession {
		best := sessions[0]
		bestScore := score(sessions[0])
		for _, snap := range sessions[1:] {
			if sc := score(snap); sc > bestScore {
				bestScore = sc
				best = snap
			}
		}
		ref := best.conv.UpdatedAt
		if ref.IsZero() {
			ref = best.conv.CreatedAt
		}
		return &NotableSession{
			Label:     label,
			SessionID: best.conv.ID,
			Value:     format(best),
			Date:      ref.Format("Jan 2, 2006"),
		}
	}

	var result []NotableSession

	if ns := pick(
		"Longest Session",
		func(snap sessionSnapshot) float64 { return snap.duration },
		func(snap sessionSnapshot) string { return fmtDuration(snap.duration) },
	); ns != nil {
		result = append(result, *ns)
	}

	if ns := pick(
		"Most Messages",
		func(snap sessionSnapshot) float64 { return float64(len(snap.conv.Messages)) },
		func(snap sessionSnapshot) string { return fmt.Sprintf("%d messages", len(snap.conv.Messages)) },
	); ns != nil {
		result = append(result, *ns)
	}

	if ns := pick(
		"Most Tool Calls",
		func(snap sessionSnapshot) float64 { return float64(snap.totalTools) },
		func(snap sessionSnapshot) string { return fmt.Sprintf("%d tool calls", snap.totalTools) },
	); ns != nil {
		result = append(result, *ns)
	}

	return result
}

func (s *InsightsService) FormatTerminal(report *InsightsReport) string {
	var sb strings.Builder

	line := func(format string, args ...interface{}) {
		fmt.Fprintf(&sb, format+"\n", args...)
	}

	line("╔══════════════════════════════════════════════════════════════╗")
	line("║               📊 Dardcor Insights                           ║")
	line("║  Last %d days · Generated %s                    ║",
		report.Days, report.GeneratedAt.Format("Jan 2, 2006 15:04"))
	line("╚══════════════════════════════════════════════════════════════╝")

	if report.Empty {
		line("")
		line("  No session data found for the last %d days.", report.Days)
		return sb.String()
	}

	ov := report.Overview

	line("")
	line("┌─ Overview ────────────────────────────────────────────────────┐")
	line("│  Sessions        %-8d    Active Days    %-8d          │", ov.TotalSessions, ov.ActiveDays)
	line("│  Messages        %-8d    Tool Calls     %-8d          │", ov.TotalMessages, ov.TotalToolCalls)
	line("│  Input Tokens    %-8d    Output Tokens  %-8d          │", ov.TotalInputTokens, ov.TotalOutputTokens)
	line("│  Total Tokens    %-8d    Est. Cost      $%-7.4f          │", ov.TotalTokens, ov.EstimatedCost)
	line("│  Avg Duration    %-16s Avg Msgs/Session %-5.1f      │",
		fmtDuration(ov.AvgSessionDuration), ov.AvgMessagesPerSession)
	if !ov.DateRangeStart.IsZero() {
		line("│  Date Range      %s → %s                    │",
			ov.DateRangeStart.Format("Jan 2"), ov.DateRangeEnd.Format("Jan 2, 2006"))
	}
	line("└───────────────────────────────────────────────────────────────┘")

	if len(report.Models) > 0 {
		line("")
		line("┌─ Models Used ─────────────────────────────────────────────────┐")
		line("│  %-22s  %6s  %8s  %8s  %8s  │",
			"Model", "Reqs", "In Tok", "Out Tok", "Cost $")
		line("│  %-22s  %6s  %8s  %8s  %8s  │",
			strings.Repeat("─", 22), strings.Repeat("─", 6),
			strings.Repeat("─", 8), strings.Repeat("─", 8), strings.Repeat("─", 8))
		for _, mb := range report.Models {
			line("│  %-22s  %6d  %8d  %8d  %8.4f  │",
				insightsTruncate(mb.Model, 22), mb.Sessions, mb.InputTokens, mb.OutputTokens, mb.Cost)
		}
		line("└───────────────────────────────────────────────────────────────┘")
	}

	if len(report.Tools) > 0 {
		line("")
		line("┌─ Top Tools ───────────────────────────────────────────────────┐")
		bars := toolBars(report.Tools, 30)
		for i, tb := range report.Tools {
			if i >= 10 {
				break
			}
			bar := ""
			if i < len(bars) {
				bar = bars[i]
			}
			line("│  %-20s  %s  %4d  (%4.1f%%)  │",
				insightsTruncate(tb.Tool, 20), bar, tb.Count, tb.Percentage)
		}
		line("└───────────────────────────────────────────────────────────────┘")
	}

	if act := report.Activity; act != nil {
		line("")
		line("┌─ Activity Patterns ───────────────────────────────────────────┐")

		dayValues := make([]int, len(act.ByDay))
		for i, d := range act.ByDay {
			dayValues[i] = d.Count
		}
		dayBars := barChart(dayValues, 28)
		line("│  Day of Week:                                                 │")
		for i, d := range act.ByDay {
			bar := ""
			if i < len(dayBars) {
				bar = dayBars[i]
			}
			line("│    %-3s  %s  %3d  │", d.Day, padBar(bar, 28), d.Count)
		}

		line("│                                                               │")

		line("│  Hour of Day (every 3h):                                      │")
		hourValues := make([]int, 24)
		for i, h := range act.ByHour {
			hourValues[i] = h.Count
		}
		hourBarsAll := barChart(hourValues, 20)
		for h := 0; h < 24; h += 3 {
			bar := ""
			if h < len(hourBarsAll) {
				bar = hourBarsAll[h]
			}
			label := fmt.Sprintf("%02d:00", h)
			line("│    %s  %s  %3d  │", label, padBar(bar, 20), hourValues[h])
		}

		line("│                                                               │")
		if act.BusiestDay != nil {
			line("│  Busiest Day:  %-10s (%d sessions)                    │",
				act.BusiestDay.Day, act.BusiestDay.Count)
		}
		if act.BusiestHour != nil {
			line("│  Peak Hour:    %-5s  (%d sessions)                        │",
				hourLabel12(act.BusiestHour.Hour), act.BusiestHour.Count)
		}
		line("│  Active Days:  %-4d    Max Streak: %d days                   │",
			act.ActiveDays, act.MaxStreak)
		line("└───────────────────────────────────────────────────────────────┘")
	}

	if len(report.TopSessions) > 0 {
		line("")
		line("┌─ Notable Sessions ────────────────────────────────────────────┐")
		for _, ns := range report.TopSessions {
			line("│  %-18s  %-30s  %s  │",
				ns.Label, insightsTruncate(ns.Value, 30), ns.Date)
		}
		line("└───────────────────────────────────────────────────────────────┘")
	}

	line("")
	return sb.String()
}

func (s *InsightsService) FormatCompact(report *InsightsReport) string {
	var sb strings.Builder

	w := func(format string, args ...interface{}) {
		fmt.Fprintf(&sb, format+"\n", args...)
	}

	w("## 📊 Dardcor Insights — Last %d Days", report.Days)
	w("_Generated %s_", report.GeneratedAt.Format("Jan 2, 2006 15:04"))
	w("")

	if report.Empty {
		w("_No session data found for the requested period._")
		return sb.String()
	}

	ov := report.Overview
	w("### Overview")
	w("**Sessions:** %d  |  **Active Days:** %d  |  **Tool Calls:** %d",
		ov.TotalSessions, ov.ActiveDays, ov.TotalToolCalls)
	w("**Messages:** %d  |  **Avg/Session:** %.1f  |  **Avg Duration:** %s",
		ov.TotalMessages, ov.AvgMessagesPerSession, fmtDuration(ov.AvgSessionDuration))
	w("**Tokens:** %d in / %d out  |  **Est. Cost:** $%.4f",
		ov.TotalInputTokens, ov.TotalOutputTokens, ov.EstimatedCost)
	w("")

	if len(report.Models) > 0 {
		w("### Models Used")
		for _, mb := range report.Models {
			w("- **%s** — %d reqs, %d tokens, $%.4f", mb.Model, mb.Sessions, mb.TotalTokens, mb.Cost)
		}
		w("")
	}

	if len(report.Tools) > 0 {
		w("### Top Tools")
		for i, tb := range report.Tools {
			if i >= 5 {
				break
			}
			w("- **%s** — %d calls (%.1f%%)", tb.Tool, tb.Count, tb.Percentage)
		}
		w("")
	}

	if act := report.Activity; act != nil {
		w("### Activity")
		if act.BusiestDay != nil {
			w("**Busiest Day:** %s (%d sessions)", act.BusiestDay.Day, act.BusiestDay.Count)
		}
		if act.BusiestHour != nil {
			w("**Peak Hour:** %s (%d sessions)", hourLabel12(act.BusiestHour.Hour), act.BusiestHour.Count)
		}
		w("**Max Streak:** %d days", act.MaxStreak)
		w("")
	}

	if len(report.TopSessions) > 0 {
		w("### Notable Sessions")
		for _, ns := range report.TopSessions {
			w("- **%s:** %s (%s)", ns.Label, ns.Value, ns.Date)
		}
		w("")
	}

	return sb.String()
}

func fmtDuration(seconds float64) string {
	s := int(math.Round(seconds))
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	if s < 3600 {
		return fmt.Sprintf("%dm %ds", s/60, s%60)
	}
	h := s / 3600
	m := (s % 3600) / 60
	return fmt.Sprintf("%dh %dm", h, m)
}

func barChart(values []int, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 20
	}
	maxVal := 0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	bars := make([]string, len(values))
	for i, v := range values {
		if maxVal == 0 {
			bars[i] = strings.Repeat("░", maxWidth)
			continue
		}
		filled := int(math.Round(float64(v) / float64(maxVal) * float64(maxWidth)))
		empty := maxWidth - filled
		bars[i] = strings.Repeat("█", filled) + strings.Repeat("░", empty)
	}
	return bars
}

func toolBars(tools []ToolBreakdown, maxWidth int) []string {
	bars := make([]string, len(tools))
	for i, tb := range tools {
		filled := int(math.Round(tb.Percentage / 100.0 * float64(maxWidth)))
		empty := maxWidth - filled
		if empty < 0 {
			empty = 0
		}
		bars[i] = strings.Repeat("█", filled) + strings.Repeat("░", empty)
	}
	return bars
}

func padBar(bar string, width int) string {
	runes := []rune(bar)
	if len(runes) >= width {
		return string(runes[:width])
	}
	return bar + strings.Repeat(" ", width-len(runes))
}

func hourLabel12(hour int) string {
	switch {
	case hour == 0:
		return "12AM"
	case hour < 12:
		return fmt.Sprintf("%dAM", hour)
	case hour == 12:
		return "12PM"
	default:
		return fmt.Sprintf("%dPM", hour-12)
	}
}

func insightsTruncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}
