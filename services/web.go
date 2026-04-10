package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	reScript     = regexp.MustCompile(`<script[\s\S]*?</script>`)
	reStyle      = regexp.MustCompile(`<style[\s\S]*?</style>`)
	reTags       = regexp.MustCompile(`<[^>]+>`)
	reWhitespace = regexp.MustCompile(`[^\S\n]+`)
	reBlankLines = regexp.MustCompile(`\n{3,}`)
)

type WebService struct {
	client *http.Client
}

func NewWebService() *WebService {
	return &WebService{
		client: &http.Client{
			Timeout: 60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (ws *WebService) Fetch(rawURL string, maxChars int) (string, error) {
	if maxChars <= 0 {
		maxChars = 50000
	}

	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://" + rawURL
	}

	_, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := ws.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	content := extractText(string(body))

	if len(content) > maxChars {
		content = content[:maxChars] + "\n... (truncated)"
	}

	return content, nil
}

func (ws *WebService) SearchDDG(query string, count int) (string, error) {
	if count <= 0 {
		count = 5
	}

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := ws.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read search response: %w", err)
	}

	reDDGLink := regexp.MustCompile(`<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>([\s\S]*?)</a>`)
	reDDGSnippet := regexp.MustCompile(`<a class="result__snippet[^"]*".*?>([\s\S]*?)</a>`)

	matches := reDDGLink.FindAllStringSubmatch(string(body), count+5)
	snippetMatches := reDDGSnippet.FindAllStringSubmatch(string(body), count+5)

	if len(matches) == 0 {
		return fmt.Sprintf("No results found for: %s", query), nil
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Results for: %s", query))

	maxItems := len(matches)
	if maxItems > count {
		maxItems = count
	}

	for i := 0; i < maxItems; i++ {
		urlStr := matches[i][1]
		title := stripHTMLTags(matches[i][2])
		title = strings.TrimSpace(title)

		if strings.Contains(urlStr, "uddg=") {
			if u, uErr := url.QueryUnescape(urlStr); uErr == nil {
				if _, after, ok := strings.Cut(u, "uddg="); ok {
					urlStr = after
				}
			}
		}

		lines = append(lines, fmt.Sprintf("%d. %s\n   %s", i+1, title, urlStr))

		if i < len(snippetMatches) {
			snippet := stripHTMLTags(snippetMatches[i][1])
			snippet = strings.TrimSpace(snippet)
			if snippet != "" {
				lines = append(lines, fmt.Sprintf("   %s", snippet))
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}

func extractText(html string) string {
	text := reScript.ReplaceAllString(html, "")
	text = reStyle.ReplaceAllString(text, "")
	text = reTags.ReplaceAllString(text, "")
	text = reWhitespace.ReplaceAllString(text, " ")
	text = reBlankLines.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func stripHTMLTags(content string) string {
	return reTags.ReplaceAllString(content, "")
}
