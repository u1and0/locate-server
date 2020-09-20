package locater

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

type (
	// History : logfileから読み込んだ検索キーワードと検索時刻
	History struct {
		KeyWord  string    // 検索キーワード
		Datetime time.Time // 検索時刻
	}
)

// LogWord extract search word from logfile
func LogWord(logfile string) (words []History, err error) {
	fp, err := os.Open(logfile)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	for {
		var (
			loc  Locater
			line []byte
			d    time.Time
		)
		line, _, err = reader.ReadLine()
		if err == io.EOF { // if EOF then finish func
			err = nil
			break
		}
		if err != nil {
			return
		}
		// 検索履歴の抽出・加工
		lines := string(line)
		if !strings.Contains(lines, "PUSH") && !strings.Contains(lines, "GET") {
			continue // ERROR行 INFO行を無視
		}
		// 検索エラーのない文字列だけwordsに追加する
		loc.SearchWords, loc.ExcludeWords, err = QueryParser(ExtractKeyword(lines))
		if err != nil {
			continue // Ignore QueryParser() Error
		}
		d, err = ExtractDatetime(lines)
		if err != nil {
			continue // Ignore time.Parse() Error
		}
		words = append(words, History{loc.Normalize(), d})
	}
	// words = SliceUnique(words)
	return
}

// ExtractDatetime extract search datetime from a line of `locate.log` format
//		Log		[NOTICE] 2020-07-07 06:57:27
func ExtractDatetime(s string) (time.Time, error) {
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	layout := "2006-01-02"
	s = re.FindString(s)
	return time.Parse(layout, s)
}

// ExtractKeyword extract search keyword from a line of `locate.log` format
func ExtractKeyword(s string) string {
	start := strings.Index(s, "[ ")
	end := strings.Index(s, " ]")
	if start < 0 || end < 0 { // Not Found "[ ]"
		return s
	}
	return s[start+1 : end-1]
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
