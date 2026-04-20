package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RulesService loads .dardcorrules files from the project directory and
// injects them into the agent's system prompt — inspired by MIAW-CLI rules system.
type RulesService struct {
	mu    sync.RWMutex
	cache map[string]string // directory → rules content
}

var rulesFileNames = []string{
	".dardcorrules",
	".miawrules",
	"AGENTS.md",
	"CLAUDE.md",
	".cursorrules",
}

func NewRulesService() *RulesService {
	return &RulesService{cache: make(map[string]string)}
}

// LoadRules walks up from dir until it finds a rules file or hits the root.
func (r *RulesService) LoadRules(dir string) string {
	r.mu.RLock()
	if cached, ok := r.cache[dir]; ok {
		r.mu.RUnlock()
		return cached
	}
	r.mu.RUnlock()

	content := r.searchUpward(dir)
	r.mu.Lock()
	r.cache[dir] = content
	r.mu.Unlock()
	return content
}

func (r *RulesService) searchUpward(dir string) string {
	current := dir
	for {
		for _, name := range rulesFileNames {
			path := filepath.Join(current, name)
			if data, err := os.ReadFile(path); err == nil {
				return fmt.Sprintf("\n\n[PROJECT RULES from %s]\n%s\n[END RULES]", name, string(data))
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

// InvalidateCache clears the rules cache so next call re-reads from disk.
func (r *RulesService) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]string)
}

// BuildRulesPrompt returns a formatted rules section or empty string.
func (r *RulesService) BuildRulesPrompt(workspaceDir string) string {
	rules := r.LoadRules(workspaceDir)
	if rules == "" {
		return ""
	}
	return rules
}

// GetRulesFiles returns all rules file paths found (for display).
func (r *RulesService) GetRulesFiles(dir string) []string {
	var found []string
	current := dir
	for {
		for _, name := range rulesFileNames {
			path := filepath.Join(current, name)
			if _, err := os.Stat(path); err == nil {
				found = append(found, path)
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return found
}

// InjectRules adds rules content to a system prompt.
func (r *RulesService) InjectRules(systemPrompt, workspaceDir string) string {
	rules := r.BuildRulesPrompt(workspaceDir)
	if rules == "" {
		return systemPrompt
	}
	return systemPrompt + rules
}

// GetTemplateList returns available .dardcorrules templates.
func (r *RulesService) GetTemplateList() []string {
	return []string{"react", "nextjs", "node-api", "python", "flutter", "rust"}
}

// GetTemplate returns a template content by name.
func GetRulesTemplate(name string) string {
	templates := map[string]string{
		"react": `# Dardcor Rules — React Project

## Tech Stack
- React 18+ with TypeScript
- Vite for bundling
- Tailwind CSS for styling

## Conventions
- Use functional components with hooks
- Follow React naming: PascalCase for components, camelCase for hooks
- Co-locate tests with components: Component.test.tsx
- Use React Query for server state

## Architecture
- src/components/ — reusable UI components
- src/pages/ — route-level components
- src/hooks/ — custom hooks
- src/services/ — API calls
- src/types/ — TypeScript types

## Quality
- All components must have TypeScript props interface
- No any types except in external library boundaries
- Prefer composition over inheritance
`,
		"nextjs": `# Dardcor Rules — Next.js Project

## Tech Stack
- Next.js 14+ with App Router
- TypeScript
- Tailwind CSS
- Prisma for database

## Conventions
- Use Server Components by default, "use client" only when needed
- API routes in app/api/
- Follow Next.js file conventions: page.tsx, layout.tsx, loading.tsx
- Use server actions for mutations

## Architecture
- app/ — App Router pages and layouts
- components/ — Shared UI components
- lib/ — Utilities and helpers
- prisma/ — Database schema and migrations
`,
		"node-api": `# Dardcor Rules — Node.js API

## Tech Stack
- Node.js 20+ with TypeScript
- Express or Fastify
- PostgreSQL with Drizzle ORM
- JWT authentication

## Conventions
- RESTful API design
- Validate all inputs with Zod
- Use async/await, no callbacks
- Error handler middleware for all routes

## Architecture
- src/routes/ — Route handlers
- src/middleware/ — Express middleware
- src/services/ — Business logic
- src/models/ — Data models
- src/utils/ — Helpers
`,
		"python": `# Dardcor Rules — Python Project

## Tech Stack
- Python 3.11+
- FastAPI for web APIs
- SQLAlchemy for database
- Pydantic for validation

## Conventions
- Use type hints everywhere
- Follow PEP 8 style guide
- Use dataclasses or Pydantic models
- pytest for testing

## Architecture
- src/ — Main package
- tests/ — Test suite
- docs/ — Documentation
`,
		"flutter": `# Dardcor Rules — Flutter Project

## Tech Stack
- Flutter 3+
- Dart
- Riverpod for state management
- GoRouter for navigation

## Conventions
- Use StatelessWidget when possible
- Separate business logic from UI with providers
- Widget naming: SuffixWidget, SuffixScreen
- Feature-first folder structure

## Architecture
- lib/features/ — Feature modules
- lib/core/ — Shared utilities, theme
- lib/data/ — API and local storage
`,
		"rust": `# Dardcor Rules — Rust Project

## Tech Stack
- Rust (stable)
- Tokio for async
- Axum for web (if applicable)
- Serde for serialization

## Conventions
- Use Result<T, E> for error handling, no unwrap() in production
- Prefer owned types in public APIs
- Document all public items with /////
- Use clippy and rustfmt

## Architecture
- src/main.rs — Entry point
- src/lib.rs — Library root
- src/handlers/ — Request handlers
- src/models/ — Data types
`,
	}
	t, ok := templates[strings.ToLower(name)]
	if !ok {
		return ""
	}
	return t
}
