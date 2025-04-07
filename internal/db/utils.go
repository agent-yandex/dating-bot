package db

import (
	"fmt"
	"sort"
	"strings"
)

const (
	insertTag = "insert"
	selectTag = "db"
	updateTag = "update"
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
