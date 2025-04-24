package db

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	insertTag = "insert"
	selectTag = "db"
	updateTag = "update"
	deleteTag = "delete"
)

func colNamesWithPref(cols []string, pref string) []string {
	prefCols := make([]string, len(cols))
	copy(prefCols, cols)
	sort.Strings(prefCols)
	if pref == "" {
		return prefCols
	}

	for i := range prefCols {
		if !strings.Contains(prefCols[i], ".") {
			prefCols[i] = fmt.Sprintf("%s.%s", pref, prefCols[i])
		}
	}
	return prefCols
}

func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)

	runes[0] = unicode.ToTitle(runes[0])

	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}
