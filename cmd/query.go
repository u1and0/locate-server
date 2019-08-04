package locater

import (
	"errors"
	"strings"
)

// sliceIn : 2つのslice中の重複要素を返す
func sliceIn(a, b []string) string {
	for _, e1 := range a {
		for _, e2 := range b {
			if e1 == e2 {
				return e1
			}
		}
	}
	return ""
}

// QueryParser : prefixがあるstringとないstringに分類してそれぞれのスライスで返す
func QueryParser(s string) (sn, en []string, err error) {
	// s <- "hoge my -your name"
	for _, n := range strings.Fields(s) { // -> [hoge my -your name]
		n = strings.ToLower(n)
		if strings.HasPrefix(n, "-") {
			en = append(en, strings.TrimPrefix(n, "-")) // ->[your]
		} else {
			sn = append(sn, n) // ->[hoge my name]
		}
	}
	if len([]rune(strings.Join(sn, ""))) < 2 {
		err = errors.New("検索文字数が足りません")
	}
	if e := sliceIn(sn, en); e != "" {
		message := "検索キーワードの中に無視するキーワードが入っています : "
		err = errors.New(message + e)
	}
	return
}
