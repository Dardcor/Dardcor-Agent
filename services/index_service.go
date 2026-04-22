package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"dardcor-agent/config"
)

type IndexService struct {
	filePath string
	index    *CodeIndex
}

type CodeIndex struct {
	RootPath  string        `json:"root_path"`
	Files     []IndexedFile `json:"files"`
	BuiltAt   time.Time     `json:"built_at"`
	FileCount int           `json:"file_count"`
}

type IndexedFile struct {
	Path     string `json:"path"`
	Ext      string `json:"ext"`
	Size     int64  `json:"size"`
	Lines    int    `json:"lines"`
	Preview  string `json:"preview"`
	Language string `json:"language"`
}

var textExtensions = map[string]string{
	".ts": "typescript", ".tsx": "typescript", ".js": "javascript", ".jsx": "javascript",
	".go": "go", ".py": "python", ".rs": "rust", ".java": "java",
	".c": "c", ".cpp": "cpp", ".h": "c", ".cs": "csharp",
	".rb": "ruby", ".php": "php", ".swift": "swift", ".kt": "kotlin",
	".md": "markdown", ".json": "json", ".yaml": "yaml", ".yml": "yaml",
	".toml": "toml", ".env": "env", ".sql": "sql", ".sh": "shell",
	".html": "html", ".css": "css", ".scss": "scss",
}

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "dist": true, "build": true,
	"__pycache__": true, "vendor": true, ".next": true, "target": true,
	".nuxt": true, "coverage": true, ".pytest_cache": true,
}

func NewIndexService() *IndexService {
	filePath := "database/code_index.json"
	if config.AppConfig != nil {
		filePath = filepath.Join(config.AppConfig.DataDir, "code_index.json")
	}
	svc := &IndexService{filePath: filePath}
	svc.loadIndex()
	return svc
}

func (s *IndexService) loadIndex() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var idx CodeIndex
	if json.Unmarshal(data, &idx) == nil {
		s.index = &idx
	}
}

func (s *IndexService) Build(rootPath string) (*CodeIndex, error) {
	if rootPath == "" {
		var err error
		rootPath, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	idx := &CodeIndex{
		RootPath: rootPath,
		BuiltAt:  time.Now(),
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 1024*1024 {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		lang, ok := textExtensions[ext]
		if !ok {
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		lines := strings.Count(content, "\n") + 1
		preview := content
		if len(preview) > 300 {
			preview = preview[:300]
		}

		idx.Files = append(idx.Files, IndexedFile{
			Path:     relPath,
			Ext:      ext,
			Size:     info.Size(),
			Lines:    lines,
			Preview:  preview,
			Language: lang,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	idx.FileCount = len(idx.Files)
	s.index = idx

	os.MkdirAll(filepath.Dir(s.filePath), 0755)
	data, _ := json.MarshalIndent(idx, "", "  ")
	os.WriteFile(s.filePath, data, 0644)

	return idx, nil
}

func (s *IndexService) Search(query string, maxResults int) []IndexedFile {
	if s.index == nil || query == "" {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 20
	}

	lower := strings.ToLower(query)
	var results []IndexedFile

	for _, f := range s.index.Files {
		score := 0
		lowerPath := strings.ToLower(f.Path)
		lowerPreview := strings.ToLower(f.Preview)

		if strings.Contains(lowerPath, lower) {
			score += 10
		}
		if strings.Contains(lowerPreview, lower) {
			score += 5
		}
		if score > 0 {
			results = append(results, f)
		}
		if len(results) >= maxResults*2 {
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return len(results[i].Path) < len(results[j].Path)
	})

	if len(results) > maxResults {
		results = results[:maxResults]
	}
	return results
}

func (s *IndexService) GetStatus() string {
	if s.index == nil {
		return "No index built. Use `dardcor index` to build."
	}
	return fmt.Sprintf("Index: %d files | Root: %s | Built: %s",
		s.index.FileCount,
		s.index.RootPath,
		s.index.BuiltAt.Format("2006-01-02 15:04"),
	)
}

func (s *IndexService) GetIndex() *CodeIndex {
	return s.index
}
