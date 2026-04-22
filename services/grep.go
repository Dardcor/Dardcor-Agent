package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GrepResult struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

type GrepService struct{}

func NewGrepService() *GrepService {
	return &GrepService{}
}

func (gs *GrepService) Search(rootPath string, pattern string, maxResults int, fileFilter string) ([]GrepResult, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	var re *regexp.Regexp
	var isRegex bool
	compiled, err := regexp.Compile(pattern)
	if err == nil {
		re = compiled
		isRegex = true
	}

	var results []GrepResult
	lowerPattern := strings.ToLower(pattern)

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if len(results) >= maxResults {
			return filepath.SkipAll
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "dist" || base == "vendor" || base == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Size() > 2*1024*1024 {
			return nil
		}

		if fileFilter != "" {
			ext := filepath.Ext(info.Name())
			if !strings.EqualFold(ext, "."+fileFilter) {
				return nil
			}
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if isBinaryExt(ext) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			matched := false
			if isRegex {
				matched = re.MatchString(line)
			} else {
				matched = strings.Contains(strings.ToLower(line), lowerPattern)
			}

			if matched {
				relPath, _ := filepath.Rel(rootPath, path)
				if relPath == "" {
					relPath = path
				}
				results = append(results, GrepResult{
					File:    relPath,
					Line:    lineNum,
					Content: strings.TrimSpace(line),
				})
				if len(results) >= maxResults {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	return results, err
}

func (gs *GrepService) Glob(rootPath string, pattern string, maxResults int) ([]string, error) {
	if maxResults <= 0 {
		maxResults = 100
	}

	lowerPattern := strings.ToLower(pattern)
	var matches []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if len(matches) >= maxResults {
			return filepath.SkipAll
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "dist" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		name := strings.ToLower(info.Name())
		if strings.Contains(name, lowerPattern) {
			relPath, _ := filepath.Rel(rootPath, path)
			if relPath == "" {
				relPath = path
			}
			matches = append(matches, relPath)
		}

		return nil
	})

	return matches, err
}

func (gs *GrepService) CountLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	return count, nil
}

func isBinaryExt(ext string) bool {
	binExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true, ".webp": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wav": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
		".bin": true, ".dat": true, ".db": true, ".sqlite": true,
	}
	return binExts[ext]
}