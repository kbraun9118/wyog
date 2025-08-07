package repository

import (
	"fmt"
	"path/filepath"
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

func (i *Ignores) CheckIgnore(path string) (bool, error) {
	if filepath.IsAbs(path) {
		return false, fmt.Errorf("This function requires path to be relative to the repo's root")
	}

	result := checkIgnoreScoped(i.Scoped, path)
	if result != nil {
		return *result, nil
	}

	return checkIgnoreAbsolute(i.Absolute, path), nil
}

func checkIgnore1(rules []IgnoreRule, path string) *bool {
	var result *bool
	for _, rule := range rules {
		if ignore.CompileIgnoreLines(rule.Pattern).MatchesPath(path) {
			result = &rule.Exclude
		}
	}
	return result
}

func checkIgnoreScoped(rulesMap map[string][]IgnoreRule, path string) *bool {
	parent := filepath.Dir(path)
	for {
		if rules, ok := rulesMap[parent]; ok {
			result := checkIgnore1(rules, path)
			if result != nil {
				return result
			}
		}
		parentParent := filepath.Dir(parent)
		if parent == "" || parent == parentParent {
			break
		}
		parent = parentParent
	}

	return nil
}

func checkIgnoreAbsolute(rules []IgnoreRule, path string) bool {
	result := checkIgnore1(rules, path)
	if result != nil {
		return *result
	}

	return false
}
