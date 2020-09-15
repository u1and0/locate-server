package locater

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const (
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
)

// LogWord extract search word from LOGFILE
func LogWord() (words []string, err error) {
	fp, err := os.Open(LOGFILE)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	for {
		var (
			loc  Locater
			line []byte
		)
		line, _, err = reader.ReadLine()
		if err == io.EOF { // if EOF then finish func
			err = nil
			break
		}
		if err != nil {
			return
		}
		lines := string(line)
		if start := strings.Index(lines, "[ "); start > -1 { // Not found "[  ]"
			end := strings.Index(lines, " ]")
			s := lines[start+1 : end-1]
			// 検索エラーのない文字列だけwordsに追加する
			if loc.SearchWords, loc.ExcludeWords, err = QueryParser(s); err == nil {
				words = append(words, loc.Normalize())
			}
		}
	}
	words = SliceUnique(words)
	return
}

// SliceUnique prune duplicate words in slice
func SliceUnique(target []string) (unique []string) {
	m := map[string]bool{}
	for _, v := range target {
		if !m[v] {
			m[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}
