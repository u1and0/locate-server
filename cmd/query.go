package locater

import (
	"errors"
	"strings"
	"unicode"
)

type (
	// Query : URL で指定されてくるAPIオプション
	Query struct {
		Q       string `form:"q"`       // 検索キーワード,除外キーワードクエリ
		Logging bool   `form:"logging"` // LOGFILEに検索記録を残すか default ture
		// 検索結果上限数
		// LimitをUintにしなかったのは、head の-nオプションが負の整数も受け付けるため。
		// 負の整数を受け付けた場合は、-n=-1と同じく、制限なしに検索結果を出力する
		Limit int `form:"limit"`
	}
)

// New : Query constructor
// Default value Logging: ture <= always log search query
//									if ommited URL request &logging
// Default value Limit: -1 <= dump all result
//									if ommited URL request &limit
func (q *Query) New() *Query {
	return &Query{Logging: true, Limit: -1}
}

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
func QueryParser(ss string) (sn, en []string, err error) {
	// s <- "hoge my -your name\D"
	// バックスラッシュの後の1文字以外は小文字化
	for _, s := range strings.Fields(ss) { // -> [hoge my -your name\D]
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
	if e := sliceIn(sn, en); e != "" {
		message := "検索キーワードの中に無視するキーワードが入っています : "
		err = errors.New(message + e)
		return
	}
	return
}

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
