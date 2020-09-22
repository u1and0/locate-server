package locater

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type (
	// History : logfileから読み込んだ検索キーワードと検索時刻
	History map[string][]time.Time
	// Keyword  string    // 検索キーワード
	// Datetime []time.Time // 検索時刻
	// }
	// Histories []History

	// Frecency : A coined word of "frequently" + "recency"
	Frecency struct {
		Word  string
		Score int
	}
	// FrecencyList : List of Frecency sorted by Frecency.Score
	FrecencyList []Frecency
)

// LogWord extract search word from logfile
func LogWord(logfile string) (History, error) {
	history := make(History, 100)
	fp, err := os.Open(logfile)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	for {
		var (
			loc   Locater
			line  []byte
			event time.Time
		)
		line, _, err = reader.ReadLine()
		if err == io.EOF { // if EOF then finish func
			err = nil
			break
		}
		if err != nil {
			return history, err
		}
		// 検索履歴の抽出・加工
		lines := string(line)
		if !strings.Contains(lines, "PUSH") && !strings.Contains(lines, "GET") {
			continue // ERROR行 INFO行を無視
		}
		// 検索エラーのない文字列だけfrecencyに追加する
		loc.SearchWords, loc.ExcludeWords, err = QueryParser(ExtractKeyword(lines))
		if err != nil {
			continue // Ignore QueryParser() Error
		}
		word := loc.Normalize()
		event, err = ExtractDatetime(lines)
		if err != nil {
			continue // Ignore time.Parse() Error
		}
		history[word] = append(history[word], event)
	}
	// words = SliceUnique(words)
	return history, err
}

// ExtractDatetime extract search datetime from a line of `locate.log` format
//		Log		[NOTICE] 2020-07-07 06:57:27
func ExtractDatetime(s string) (time.Time, error) {
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	layout := "2006-01-02 15:04:05"
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

// Scoring : 日時から頻出度を算出する
func Scoring(t time.Time) int { // map[string]int{
	since := time.Since(t).Hours()
	switch {
	case since < 6:
		return 16
	case since < 24:
		return 8
	case since < 24*7:
		return 4
	case since < 24*28:
		return 2
	default:
		return 1
	}
}

//ScoreSum : 履歴マップの検索日時リストからスコア合計を算出する
func ScoreSum(tl []time.Time) (score int) {
	for _, t := range tl {
		score += Scoring(t)
	}
	return
}

// RankByScore : 履歴から頻出度リストを生成する
func (history History) RankByScore() FrecencyList {
	var i int
	l := make(FrecencyList, len(history))
	for k, v := range history {
		l[i] = Frecency{k, ScoreSum(v)}
		i++
	}
	sort.Sort(l)
	return l
}

func (p FrecencyList) Len() int           { return len(p) }
func (p FrecencyList) Less(i, j int) bool { return p[i].Score > p[j].Score }
func (p FrecencyList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
