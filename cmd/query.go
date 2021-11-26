package locater

import (
	"errors"
	"strings"
	"unicode"
)

// ToLowerExceptFirst : To lower except first of runes
func ToLowerExceptFirst(s string) string {
	runes := []rune(s)
	for i, r := range runes { // to LOWER except first position
		if i != 0 {
			runes[i] = unicode.ToLower(r)
		}
	}
	return string(runes)
}

// ToLowerExceptAll : ToLower string Except specific rune for whole words
func ToLowerExceptAll(s string, r rune) string {
	st := strings.Split(s, "\\")
	for i, si := range st {
		st[i] = ToLowerExceptFirst(si)
	}
	return strings.Join(st, "\\")
}

// QueryParser : prefixがあるstringとないstringに分類してそれぞれのスライスで返す
func QueryParser(query string) (sn, en []string, err error) {
	// s <- "hoge my -your name\D"
	// バックスラッシュの後の1文字以外は小文字化
	for _, s := range strings.Fields(query) { // -> [hoge my -your name\D]
		if strings.Contains(s, "\\") {
			s = ToLowerExceptAll(s, '\\')
		} else {
			s = strings.ToLower(s)
		}

		// 文字列頭に"-"がついていたらExcludeWords, そうでなければSearchWords
		if strings.HasPrefix(s, "-") {
			en = append(en, strings.TrimPrefix(s, "-")) // ->[your]
		} else {
			sn = append(sn, s) // ->[hoge my name]
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
		return
	}
	// snとenに重複する語が入っていたらerror
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
		return
	}
	return
}
