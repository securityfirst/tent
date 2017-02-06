package utils

import (
	"regexp"
	"strings"
)

var (
	cleanAlphaNum = regexp.MustCompile("[^[:alpha:]\\s\\d_]")
	spaceTrim     = regexp.MustCompile("\\s+")
)

func MakeId(v string) string {
	v = cleanAlphaNum.ReplaceAllString(v, " ")
	v = spaceTrim.ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")
	v = strings.ToLower(v)
	return v
}
