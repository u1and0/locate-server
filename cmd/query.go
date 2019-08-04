package locater

import (
	"errors"
	"strings"
)

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
	// 各検索語のどれかが2文字以上ならnot error
	if func() bool {
		for _, s := range sn {
			if len([]rune(s)) > 1 {
				return false
			}
		}
		return true
	}() {
		message := "検索文字数が足りません : "
		err = errors.New(message + strings.Join(sn, " "))
	}
	// snとenに重複する語が入っているたらerror
	if e := func() string {
		for _, s := range sn {
			for _, e := range en {
				if s == e {
					return s
				}
			}
		}
		return ""
	}(); e != "" {
		message := "検索キーワードの中に無視するキーワードが入っています : "
		err = errors.New(message + e)
	}
	return
}
