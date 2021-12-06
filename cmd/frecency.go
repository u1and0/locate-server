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

	// Frecency : A coined word of "frequently" + "recency"
	Frecency struct {
		Word  string `json:"word"`
		Score int    `json:"score"`
	}
	// SearchHistory : List of Frecency sorted by Frecency.Score
	SearchHistory []Frecency
)

// LogWord extract search word from logfile
func LogWord(logfile string) (History, error) {
	history := make(History, 100)
	fp, err := os.Open(logfile)
	if err != nil {
		return history, err
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	for {
		var (
			l     Locater
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
		sw, ew, err := QueryParser(ExtractKeyword(lines))
		if err != nil {
			continue // Ignore QueryParser() Error
		}
		l.SearchWords, l.ExcludeWords = sw, ew
		word := l.Normalize()
		event, err = ExtractDatetime(lines)
		if err != nil {
			continue // Ignore time.Parse() Error
		}
		history[word] = append(history[word], event)
	}
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

// Scoring : 日時から頻出度を算出する
func Scoring(t time.Time) int {
	since := time.Since(t).Hours()
	switch {
	case since < 6:
		return 32
	case since < 24:
		return 16
	case since < 24*7:
		return 8
	case since < 24*14:
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
func (history History) RankByScore() SearchHistory {
	var i int
	l := make(SearchHistory, len(history))
	for k, v := range history {
		l[i] = Frecency{k, ScoreSum(v)}
		i++
	}
	sort.Sort(l)
	return l
}

func (fl SearchHistory) Len() int           { return len(fl) }
func (fl SearchHistory) Less(i, j int) bool { return fl[i].Score > fl[j].Score }
func (fl SearchHistory) Swap(i, j int)      { fl[i], fl[j] = fl[j], fl[i] }

// Datalist throw list of searched words sorted by score
func Datalist(f string) (SearchHistory, error) {
	historymap, err := LogWord(f)
	wordList := historymap.RankByScore()
	return wordList, err
}
