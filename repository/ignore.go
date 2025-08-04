package repository

import (
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

type IgnoreRule struct {
	Pattern string
	Exclude bool
}

func gitignoreParse1(raw string) *IgnoreRule {
	raw = strings.TrimSpace(raw)

	if len(raw) == 0 || raw[0] == '#' {
		return nil
	} else if raw[0] == '!' {
		return &IgnoreRule{raw[1:], false}
	} else if raw[0] == '\\' {
		return &IgnoreRule{raw[1:], true}
	} else {
		return &IgnoreRule{raw, true}
	}
}

func gitignoreParse(lines []string) []IgnoreRule {
	ret := make([]IgnoreRule, 0)

	for _, line := range lines {
		parsed := gitignoreParse1(line)
		if parsed != nil {
			ret = append(ret, *parsed)
		}
	}

	return ret
}

type Ignores struct {
	Absolute []IgnoreRule
	Scoped   map[string][]IgnoreRule
}

func NewIgnores() Ignores {
	return Ignores{
		Absolute: make([]IgnoreRule, 0),
		Scoped:   make(map[string][]IgnoreRule),
	}
}

type IgnoreMatch int

func checkIgnore1(rules []IgnoreRule, path string) *bool {
	var result *bool
	for _, rule := range rules {
		if ignore.CompileIgnoreLines(rule.Pattern).MatchesPath(path) {
			result = &rule.Exclude
		}
	}
	return result
}
