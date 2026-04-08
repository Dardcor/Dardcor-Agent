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

