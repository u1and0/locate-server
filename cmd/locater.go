package locater

import (
	"sort"
	"strings"
)

// Locater : queryから読み取った検索ワードと無視するワード
type Locater struct {
	SearchWords, ExcludeWords []string
	Dbpath                    string
	Cap                       int
	PathSplitWin              bool
	Root                      string
	Trim                      string
}

// Normalize : SearchWordsとExcludeWordsを合わせる
// SearchWordsは小文字にする
// ExcludeWordsは小文字にした上で
// ソートして、頭に-をつける
func (l *Locater) Normalize() string {
	se := l.SearchWords
	ex := l.ExcludeWords

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
