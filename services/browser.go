package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

type BrowserService struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	page    playwright.Page
	mu      sync.Mutex
}

func NewBrowserService() *BrowserService {
	return &BrowserService{}
}

func (bs *BrowserService) EnsureBrowser() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.page != nil {
		return nil
	}

	pw, err := playwright.Run()
	if err != nil {
		// Only install if run fails (likely missing drivers)
		fmt.Printf("[BrowserService] Initializing Playwright drivers (this may take a moment)...")
		if errImg := playwright.Install(); errImg != nil {
			return fmt.Errorf("could not install playwright: %w", errImg)
		}
		pw, err = playwright.Run()
		if err != nil {
			return fmt.Errorf("could not start playwright after installation: %w", err)
		}
	}
	bs.pw = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("could not launch browser: %w", err)
	}
	bs.browser = browser

	context, err := browser.NewContext()
	if err != nil {
		return fmt.Errorf("could not create browser context: %w", err)
	}
	bs.context = context

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}
	bs.page = page

	return nil
}

func (bs *BrowserService) Navigate(url string) (string, error) {
	if err := bs.EnsureBrowser(); err != nil {
		return "", err
	}
	_, err := bs.page.Goto(url)
	if err != nil {
		return "", err
	}
	title, _ := bs.page.Title()
	return fmt.Sprintf("Navigated to %s (Title: %s)", url, title), nil
}

func (bs *BrowserService) Click(selector string) (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}
	err := bs.page.Click(selector)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Clicked on %s", selector), nil
}

func (bs *BrowserService) Type(selector, text string) (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}
	err := bs.page.Fill(selector, text)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Typed '%s' into %s", text, selector), nil
}

func (bs *BrowserService) Screenshot(dataDir string) (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}

	path := filepath.Join(dataDir, "screenshots")
	os.MkdirAll(path, 0755)

	filename := fmt.Sprintf("screenshot_%d.png", time.Now().Unix())
	fullPath := filepath.Join(path, filename)

	_, err := bs.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(fullPath),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Screenshot saved to %s", filename), nil
}

func (bs *BrowserService) Scroll(direction string) (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}
	if direction == "down" {
		_, err := bs.page.Evaluate("window.scrollBy(0, window.innerHeight)")
		if err != nil {
			return "", err
		}
		return "Scrolled down", nil
	}
	_, err := bs.page.Evaluate("window.scrollBy(0, -window.innerHeight)")
	if err != nil {
		return "", err
	}
	return "Scrolled up", nil
}

func (bs *BrowserService) Wait(ms int) (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return fmt.Sprintf("Waited for %dms", ms), nil
}

func (bs *BrowserService) Back() (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}
	_, err := bs.page.GoBack()
	if err != nil {
		return "", err
	}
	return "Went back", nil
}

func (bs *BrowserService) GetDOM() (string, error) {
	if bs.page == nil {
		return "", fmt.Errorf("browser not started")
	}
	content, err := bs.page.Content()
	if err != nil {
		return "", err
	}
	if len(content) > 10000 {
		content = content[:10000] + "... (truncated)"
	}
	return content, nil
}

func (bs *BrowserService) Close() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.browser != nil {
		err := bs.browser.Close()
		bs.page = nil
		bs.browser = nil
		bs.pw.Stop()
		return err
	}
	return nil
}
