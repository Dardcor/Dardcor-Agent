package services

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type FileSafetyService struct {
	safeWriteRoot       string
	writeDeniedPaths    map[string]bool
	writeDeniedPrefixes []string
}

func NewFileSafetyService() *FileSafetyService {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	svc := &FileSafetyService{
		safeWriteRoot:       os.Getenv("DARDCOR_WRITE_SAFE_ROOT"),
		writeDeniedPaths:    BuildWriteDeniedPaths(home),
		writeDeniedPrefixes: BuildWriteDeniedPrefixes(home),
	}
	return svc
}

func BuildWriteDeniedPaths(home string) map[string]bool {
	denied := map[string]bool{}

	add := func(p string) {
		if p != "" {
			denied[filepath.Clean(p)] = true
		}
	}

	if home != "" {
		add(filepath.Join(home, ".ssh", "authorized_keys"))
		add(filepath.Join(home, ".ssh", "id_rsa"))
		add(filepath.Join(home, ".ssh", "id_ed25519"))
		add(filepath.Join(home, ".ssh", "config"))

		add(filepath.Join(home, ".dardcor", ".env"))

		add(filepath.Join(home, ".bashrc"))
		add(filepath.Join(home, ".zshrc"))
		add(filepath.Join(home, ".profile"))
		add(filepath.Join(home, ".bash_profile"))
		add(filepath.Join(home, ".zprofile"))

		add(filepath.Join(home, ".netrc"))
		add(filepath.Join(home, ".pgpass"))
		add(filepath.Join(home, ".npmrc"))
		add(filepath.Join(home, ".pypirc"))
	}

	if runtime.GOOS != "windows" {
		add("/etc/sudoers")
		add("/etc/passwd")
		add("/etc/shadow")
	} else {
		userProfile := os.Getenv("USERPROFILE")
		if userProfile != "" {
			add(filepath.Join(userProfile, ".ssh", "authorized_keys"))
			add(filepath.Join(userProfile, ".ssh", "id_rsa"))
			add(filepath.Join(userProfile, ".ssh", "id_ed25519"))
			add(filepath.Join(userProfile, ".ssh", "config"))
			add(filepath.Join(userProfile, ".dardcor", ".env"))
		}
		add(filepath.Join(os.Getenv("SystemRoot"), "System32", "config", "SAM"))
	}

	return denied
}

func BuildWriteDeniedPrefixes(home string) []string {
	prefixes := []string{}

	addDir := func(p string) {
		if p == "" {
			return
		}
		p = filepath.Clean(p) + string(filepath.Separator)
		prefixes = append(prefixes, p)
	}

	if home != "" {
		addDir(filepath.Join(home, ".ssh"))
		addDir(filepath.Join(home, ".aws"))
		addDir(filepath.Join(home, ".gnupg"))
		addDir(filepath.Join(home, ".kube"))
		addDir(filepath.Join(home, ".docker"))
		addDir(filepath.Join(home, ".azure"))
		addDir(filepath.Join(home, ".config", "gh"))
	}

	if runtime.GOOS != "windows" {
		addDir("/etc/sudoers.d")
		addDir("/etc/systemd")
	}

	return prefixes
}

func (s *FileSafetyService) IsWriteDenied(path string) bool {
	resolved := resolvePath(path)

	if s.writeDeniedPaths[filepath.Clean(resolved)] {
		return true
	}

	for _, prefix := range s.writeDeniedPrefixes {
		if strings.HasPrefix(resolved, prefix) {
			return true
		}
	}

	if s.safeWriteRoot != "" {
		safeRoot := filepath.Clean(s.safeWriteRoot) + string(filepath.Separator)
		if !strings.HasPrefix(resolved, safeRoot) {
			return true
		}
	}

	return false
}

func (s *FileSafetyService) GetReadBlockError(path string) *string {
	resolved := resolvePath(path)

	home, _ := os.UserHomeDir()

	blockedCachePaths := []string{}
	if home != "" {
		blockedCachePaths = append(blockedCachePaths,
			filepath.Join(home, ".dardcor", "skills", "index-cache.json"),
			filepath.Join(home, ".dardcor", "skills", "index-cache"),
		)
	}

	for _, blocked := range blockedCachePaths {
		if filepath.Clean(resolved) == filepath.Clean(blocked) {
			msg := fmt.Sprintf(
				"reading %q is not permitted: this file is an internal Dardcor skills hub index cache",
				path,
			)
			return &msg
		}
	}

	return nil
}

func resolvePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}

	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}

	return real
}
