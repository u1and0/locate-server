package locater

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

type (
	// Query : URL で指定されてくるAPIオプション
	Query struct {
		SearchWords  []string `json:"searchWords"`  // 検索キーワード
		ExcludeWords []string `json:"excludeWords"` // 検索から取り除くキーワード
		Logging      bool     `json:"logging"`      // LOGFILEに検索記録を残すか default ture
		Limit        int      `json:"limit"`        // 検索結果上限数
	}
)

func New() *Query {
	return &Query{}
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
func (q *Query) Parser(c *gin.Context) (err error) {
	if err = q.WordParser(c.Query("q")); err != nil {
		return
	}
	if err = q.LimitParser(c.DefaultQuery("limit", "0")); err != nil {
		return
	}
	if err = q.LoggingParser(c.DefaultQuery("logging", "false")); err != nil {
		return
	}
	return
}

func (q *Query) LoggingParser(s string) error {
	// 基本的にすべての検索はログに記録する
	// http:...&logging=falseのときだけ記録しない
	q.Logging = true
	if strings.ToLower(s) == "false" {
		q.Logging = false
		return nil
	} else if strings.ToLower(s) == "false" {
		q.Logging = true
		return nil
	}
	message := "error: &logging = true or false or undef : "
	return errors.New(message + s)
}

// LimitParser : parse &limit as Int
func (q *Query) LimitParser(s string) error { // s == "0"||Int like string
	n, err := strconv.Atoi(s)
	if n < 0 {
		n = 0
	}
	q.Limit = n
	return err
}

// WordParser :
func (q *Query) WordParser(s string) error {
	var sn, en []string
	// s <- "hoge my -your name\D"
	// バックスラッシュの後の1文字以外は小文字化
	for _, s := range strings.Fields(s) { // -> [hoge my -your name\D]
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
		return errors.New(message + strings.Join(sn, " "))
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
		return errors.New(message + e)
	}
	q.SearchWords, q.ExcludeWords = sn, en
	return nil
}

// Normalize : SearchWordsとExcludeWordsを合わせる
// SearchWordsは小文字にする
// ExcludeWordsは小文字にした上で
// ソートして、頭に-をつける
func (q *Query) Normalize() string {
	se := q.SearchWords
	ex := q.ExcludeWords

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
