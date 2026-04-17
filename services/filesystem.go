package services

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"dardcor-agent/models"
)

type FileSystemService struct{}

func NewFileSystemService() *FileSystemService {
	return &FileSystemService{}
}

func (fs *FileSystemService) ListDirectory(path string) ([]models.FileInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	var files []models.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ext := ""
		if !entry.IsDir() {
			ext = filepath.Ext(entry.Name())
		}

		files = append(files, models.FileInfo{
			Name:       entry.Name(),
			Path:       filepath.Join(absPath, entry.Name()),
			Size:       info.Size(),
			IsDir:      entry.IsDir(),
			Extension:  ext,
			ModifiedAt: info.ModTime(),
			Permission: info.Mode().String(),
		})
	}

	return files, nil
}

func (fs *FileSystemService) ReadFile(path string) (*models.FileContent, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file")
	}

	if info.Size() > 50*1024*1024 {
		return nil, fmt.Errorf("file too large (max 50MB)")
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	return &models.FileContent{
		Path:     absPath,
		Content:  string(content),
		Encoding: "utf-8",
		Size:     info.Size(),
	}, nil
}

func (fs *FileSystemService) WriteFile(path string, content string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	return os.WriteFile(absPath, []byte(content), 0644)
}

func (fs *FileSystemService) DeleteFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return os.RemoveAll(absPath)
}

func (fs *FileSystemService) CreateDirectory(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	return os.MkdirAll(absPath, 0755)
}

func (fs *FileSystemService) MoveFile(src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	return os.Rename(absSrc, absDst)
}

func (fs *FileSystemService) CopyFile(src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	sourceFile, err := os.Open(absSrc)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	os.MkdirAll(filepath.Dir(absDst), 0755)

	destFile, err := os.Create(absDst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (fs *FileSystemService) SearchFiles(req models.SearchRequest) ([]models.SearchResult, error) {
	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	var results []models.SearchResult
	maxResults := 100
	query := strings.ToLower(req.Query)

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if len(results) >= maxResults {
			return filepath.SkipAll
		}

		if req.MaxDepth > 0 {
			rel, _ := filepath.Rel(absPath, path)
			depth := strings.Count(rel, string(os.PathSeparator))
			if depth > req.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if req.FileType != "" && !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if !strings.EqualFold(ext, "."+req.FileType) {
				return nil
			}
		}

		nameMatch := strings.Contains(strings.ToLower(info.Name()), query)

		if nameMatch {
			results = append(results, models.SearchResult{
				Path:  path,
				Name:  info.Name(),
				IsDir: info.IsDir(),
			})
		}

		if req.SearchContent && !info.IsDir() && info.Size() < 1024*1024 {
			if contentResults := fs.searchInFile(path, query); len(contentResults) > 0 {
				for _, cr := range contentResults {
					if len(results) >= maxResults {
						break
					}
					results = append(results, cr)
				}
			}
		}

		return nil
	})

	return results, err
}

func (fs *FileSystemService) searchInFile(path, query string) []models.SearchResult {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var results []models.SearchResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), query) {
			results = append(results, models.SearchResult{
				Path:      path,
				Name:      filepath.Base(path),
				MatchLine: lineNum,
				MatchText: strings.TrimSpace(line),
			})
			if len(results) >= 5 {
				break
			}
		}
	}

	return results
}

func (fs *FileSystemService) GetFileInfo(path string) (*models.FileInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	ext := ""
	if !info.IsDir() {
		ext = filepath.Ext(info.Name())
	}

	return &models.FileInfo{
		Name:       info.Name(),
		Path:       absPath,
		Size:       info.Size(),
		IsDir:      info.IsDir(),
		Extension:  ext,
		ModifiedAt: info.ModTime(),
		Permission: info.Mode().String(),
	}, nil
}

func (fs *FileSystemService) GetDrives() []string {
	var drives []string
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		path := string(drive) + ":\\"
		if _, err := os.Open(path); err == nil {
			drives = append(drives, path)
		}
	}
	return drives
}

func (fs *FileSystemService) GetDefaultWorkspace() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	workspacePath := filepath.Join(home, "Documents", "Dardcor-Workspace")
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return "", err
		}
	}
	return workspacePath, nil
}
func (fs *FileSystemService) Glob(root, pattern string) ([]string, error) {
	relPattern := pattern
	if !filepath.IsAbs(pattern) {
		relPattern = filepath.Join(root, pattern)
	}
	matches, err := filepath.Glob(relPattern)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, m := range matches {
		rel, err := filepath.Rel(root, m)
		if err == nil {
			results = append(results, rel)
		} else {
			results = append(results, m)
		}
	}
	return results, nil
}

func (fs *FileSystemService) EditFile(path string, startLine, endLine int, newContent string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > len(lines) {
		startLine = len(lines) + 1
	}

	newLines := strings.Split(newContent, "\n")

	var result []string
	result = append(result, lines[:startLine-1]...)
	result = append(result, newLines...)
	if endLine < len(lines) {
		result = append(result, lines[endLine:]...)
	}

	return os.WriteFile(absPath, []byte(strings.Join(result, "\n")), 0644)
}

func (fs *FileSystemService) InsertLines(path string, afterLine int, content string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	newLines := strings.Split(content, "\n")

	if afterLine < 0 {
		afterLine = 0
	}
	if afterLine > len(lines) {
		afterLine = len(lines)
	}

	var result []string
	result = append(result, lines[:afterLine]...)
	result = append(result, newLines...)
	result = append(result, lines[afterLine:]...)

	return os.WriteFile(absPath, []byte(strings.Join(result, "\n")), 0644)
}

func (fs *FileSystemService) ReadFileLines(path string, startLine, endLine int) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	if startLine < 1 {
		startLine = 1
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > len(lines) {
		return "", fmt.Errorf("start line %d exceeds file length %d", startLine, len(lines))
	}

	selected := lines[startLine-1 : endLine]
	var sb strings.Builder
	for i, line := range selected {
		sb.WriteString(fmt.Sprintf("%d: %s\n", startLine+i, line))
	}
	return sb.String(), nil
}

func (fs *FileSystemService) ReplaceInFile(path, oldText, newText string) (int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, fmt.Errorf("invalid path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return 0, fmt.Errorf("cannot read file: %w", err)
	}

	content := string(data)
	count := strings.Count(content, oldText)
	if count == 0 {
		return 0, fmt.Errorf("text not found in file")
	}

	newContent := strings.ReplaceAll(content, oldText, newText)
	err = os.WriteFile(absPath, []byte(newContent), 0644)
	return count, err
}

func (fs *FileSystemService) AppendToFile(path, content string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	dir := filepath.Dir(absPath)
	os.MkdirAll(dir, 0755)

	f, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func (fs *FileSystemService) TreeDir(path string, maxDepth int) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(absPath + "\n")

	err = fs.walkTree(&sb, absPath, "", 0, maxDepth)
	return sb.String(), err
}

func (fs *FileSystemService) walkTree(sb *strings.Builder, dir, prefix string, depth, maxDepth int) error {
	if maxDepth > 0 && depth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	filtered := make([]os.DirEntry, 0)
	for _, e := range entries {
		name := e.Name()
		if name == "node_modules" || name == ".git" || name == "dist" || name == "vendor" || name == "__pycache__" || name == ".next" {
			continue
		}
		filtered = append(filtered, e)
	}

	for i, entry := range filtered {
		isLast := i == len(filtered)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, entry.Name()))

		if entry.IsDir() {
			nextPrefix := prefix + "│   "
			if isLast {
				nextPrefix = prefix + "    "
			}
			fs.walkTree(sb, filepath.Join(dir, entry.Name()), nextPrefix, depth+1, maxDepth)
		}
	}
	return nil
}
