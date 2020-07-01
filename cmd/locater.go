package locater

import (
	"sort"
	"strings"
)

// Locater : queryから読み取った検索ワードと無視するワード
type Locater struct {
	SearchWords  []string // 検索キーワード
	ExcludeWords []string // 検索から取り除くキーワード
	Dbpath       string   // 検索対象DBパス /path/to/database:/path/to/another
	Limit        int      // 検索結果HTML表示制限数
	PathSplitWin bool     // TrueでWindowsパスセパレータを使用する
	Root         string   // 追加するドライブパス名
	Trim         string   // 削除するドライブパス名
	Process      int      // xargsによるマルチプロセス数
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
