package services

import "regexp"

type PrivacyFilter struct {
	re *regexp.Regexp
}

func NewPrivacyFilter() *PrivacyFilter {
	return &PrivacyFilter{
		re: regexp.MustCompile(`(?s)<private>.*?</private>`),
	}
}

func (f *PrivacyFilter) Filter(content string) string {
	return f.re.ReplaceAllString(content, "")
}

func (f *PrivacyFilter) ContainsPrivateTag(content string) bool {
	return f.re.MatchString(content)
}
