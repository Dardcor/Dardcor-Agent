package services

import (
	"os"
	"regexp"
	"strings"
	"sync"
)

type redactPattern struct {
	re      *regexp.Regexp
	repl    string
	capture bool
}

type RedactService struct {
	mu       sync.RWMutex
	enabled  bool
	patterns []redactPattern
}

func NewRedactService() *RedactService {
	enabled := true
	if v := os.Getenv("DARDCOR_REDACT_SECRETS"); strings.EqualFold(v, "false") || v == "0" {
		enabled = false
	}

	svc := &RedactService{enabled: enabled}
	svc.patterns = buildPatterns()
	return svc
}

func (r *RedactService) RedactSensitiveText(text string) string {
	if !r.enabled || text == "" {
		return text
	}

	r.mu.RLock()
	patterns := r.patterns
	r.mu.RUnlock()

	for _, p := range patterns {
		if p.capture {
			text = p.re.ReplaceAllStringFunc(text, func(match string) string {
				subs := p.re.FindStringSubmatch(match)
				if len(subs) < 2 {
					return match
				}
				secret := subs[1]
				masked := maskToken(secret)
				return strings.Replace(match, secret, masked, 1)
			})
		} else {
			text = p.re.ReplaceAllString(text, p.repl)
		}
	}

	return text
}

func maskToken(token string) string {
	if len(token) < 18 {
		return "***"
	}
	return token[:6] + "..." + token[len(token)-4:]
}

func buildPatterns() []redactPattern {
	var patterns []redactPattern

	apiKeyPrefixes := []string{
		`sk-[A-Za-z0-9\-_]{16,}`,
		`ghp_[A-Za-z0-9]{36,}`,
		`github_pat_[A-Za-z0-9_]{20,}`,
		`gho_[A-Za-z0-9]{36,}`,
		`ghu_[A-Za-z0-9]{36,}`,
		`ghs_[A-Za-z0-9]{36,}`,
		`ghr_[A-Za-z0-9]{36,}`,
		`xox[baprs]-[A-Za-z0-9\-]{10,}`,
		`AIza[0-9A-Za-z\-_]{35}`,
		`pplx-[A-Za-z0-9]{48,}`,
		`fal_[A-Za-z0-9\-_]{20,}`,
		`fc-[A-Za-z0-9]{32,}`,
		`bb_live_[A-Za-z0-9]{32,}`,
		`gAAAA[A-Za-z0-9\+/=]{20,}`,
		`AKIA[0-9A-Z]{16}`,
		`sk_live_[A-Za-z0-9]{24,}`,
		`sk_test_[A-Za-z0-9]{24,}`,
		`rk_live_[A-Za-z0-9]{24,}`,
		`SG\.[A-Za-z0-9\-_]{22}\.[A-Za-z0-9\-_]{43}`,
		`hf_[A-Za-z0-9]{34,}`,
		`r8_[A-Za-z0-9]{40}`,
		`npm_[A-Za-z0-9]{36}`,
		`pypi-[A-Za-z0-9\-_]{32,}`,
		`dop_v1_[A-Za-z0-9]{64,}`,
		`doo_v1_[A-Za-z0-9]{64,}`,
		`am_[A-Za-z0-9]{32,}`,
		`tvly-[A-Za-z0-9\-_]{32,}`,
		`exa_[A-Za-z0-9\-_]{32,}`,
		`gsk_[A-Za-z0-9]{52}`,
		`syt_[A-Za-z0-9\-_]{40,}`,
		`retaindb_[A-Za-z0-9\-_]{20,}`,
		`hsk-[A-Za-z0-9]{32,}`,
		`mem0_[A-Za-z0-9\-_]{20,}`,
		`brv_[A-Za-z0-9\-_]{20,}`,
	}
	for _, prefix := range apiKeyPrefixes {
		re := regexp.MustCompile(`(?i)\b(` + prefix + `)\b`)
		patterns = append(patterns, redactPattern{re: re, capture: true})
	}

	envRe := regexp.MustCompile(
		`(?i)((?:API_KEY|TOKEN|SECRET|PASSWORD|PASSWD|CREDENTIAL|AUTH)[A-Z0-9_]*\s*=\s*)([^\s\r\n&"'` + "`" + `]{4,})`,
	)
	patterns = append(patterns, redactPattern{
		re:      envRe,
		repl:    "${1}***",
		capture: false,
	})

	jsonFieldRe := regexp.MustCompile(
		`(?i)("(?:api_?key|token|secret|password|passwd|credential|auth|authorization|access_token|refresh_token|client_secret|private_key|api_token|bearer)"\s*:\s*")([^"]{4,})("?)`,
	)
	patterns = append(patterns, redactPattern{
		re:      jsonFieldRe,
		repl:    "${1}***${3}",
		capture: false,
	})

	authHeaderRe := regexp.MustCompile(
		`(?i)(Authorization\s*:\s*(?:Bearer|Basic|Token|ApiKey|Api-Key)\s+)([A-Za-z0-9\-_\.=+/]{8,})`,
	)
	patterns = append(patterns, redactPattern{
		re:      authHeaderRe,
		repl:    "${1}***",
		capture: false,
	})

	telegramRe := regexp.MustCompile(
		`(?i)\b(bot)?(\d{8,12}:[A-Za-z0-9\-_]{35})\b`,
	)
	patterns = append(patterns, redactPattern{
		re:      telegramRe,
		capture: false,
		repl:    "${1}***",
	})

	privateKeyRe := regexp.MustCompile(
		`(?s)(-----BEGIN [A-Z ]* ?PRIVATE KEY-----)(.+?)(-----END [A-Z ]* ?PRIVATE KEY-----)`,
	)
	patterns = append(patterns, redactPattern{
		re:      privateKeyRe,
		repl:    "${1}[REDACTED]${3}",
		capture: false,
	})

	dbConnRe := regexp.MustCompile(
		`(?i)((?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?|redis(?:s)?|amqps?)://[^:@\s]+:)([^@\s]{4,})(@)`,
	)
	patterns = append(patterns, redactPattern{
		re:      dbConnRe,
		repl:    "${1}***${3}",
		capture: false,
	})

	jwtRe := regexp.MustCompile(
		`\b(eyJ[A-Za-z0-9\-_]+\.eyJ[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+)\b`,
	)
	patterns = append(patterns, redactPattern{re: jwtRe, capture: true})

	urlQueryRe := regexp.MustCompile(
		`(?i)([?&](?:access_token|refresh_token|api_key|apikey|password|passwd|secret|token|credential|auth|client_secret|private_key)=)([^&\s#"'` + "`" + `]{4,})`,
	)
	patterns = append(patterns, redactPattern{
		re:      urlQueryRe,
		repl:    "${1}***",
		capture: false,
	})

	urlUserinfoRe := regexp.MustCompile(
		`(?i)(https?://[^:@\s]+:)([^@\s]{4,})(@)`,
	)
	patterns = append(patterns, redactPattern{
		re:      urlUserinfoRe,
		repl:    "${1}***${3}",
		capture: false,
	})

	discordMentionRe := regexp.MustCompile(
		`<@!?(\d{17,19})>`,
	)
	patterns = append(patterns, redactPattern{
		re:      discordMentionRe,
		repl:    "<@[REDACTED]>",
		capture: false,
	})

	phoneRe := regexp.MustCompile(
		`\+([1-9]\d{6,14})\b`,
	)
	patterns = append(patterns, redactPattern{
		re:      phoneRe,
		repl:    "+***",
		capture: false,
	})

	formBodyRe := regexp.MustCompile(
		`(?i)((?:^|&)(?:access_token|refresh_token|api_key|apikey|password|passwd|secret|token|credential|auth|client_secret|private_key)=)([^&\s]{4,})`,
	)
	patterns = append(patterns, redactPattern{
		re:      formBodyRe,
		repl:    "${1}***",
		capture: false,
	})

	return patterns
}
