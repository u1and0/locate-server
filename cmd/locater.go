package cmd

import (
	"sort"
	"strings"
)

// Locater : queryから読み取った検索ワードと無視するワード
type Locater struct {
	SearchWords, ExcludeWords []string
}

// Normalize : SearchWordsとExcludeWordsを合わせる。ExcludeWordsに"-"のprefixを入れる
func (l *Locater) Normalize() string {
	se := l.SearchWords
	ex := l.ExcludeWords

	// Convert lower for normalizing
	for i, e := range se {
		se[i] = strings.ToLower(e)
	}
	for i, e := range ex {
		ex[i] = strings.ToLower(e)
	}
	// Sort
	sort.Slice(ex, func(i, j int) bool { return ex[i] < ex[j] })
	// Add prefix "-"
	strs := append(se, func() (d []string) {
		for _, ex := range ex {
			d = append(d, "-"+ex)
		}
		return
	}()...)
	return strings.Join(strs, " ")
}
