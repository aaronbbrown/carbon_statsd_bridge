package main

import (
	"regexp"
)

type transform struct {
	Match   string
	Replace string
	Regexp  *regexp.Regexp
}

func compileTransforms(transforms []transform) []transform {
	var result []transform
	for _, t := range transforms {
		t.Regexp = regexp.MustCompile(t.Match)
		result = append(result, t)
	}
	return result
}
