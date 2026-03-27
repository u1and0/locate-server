package locater

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	api "github.com/u1and0/locate-server/cmd/api"
)

type (
	// historyMap : logfileから読み込んだ検索キーワードと検索時刻
	historyMap map[string][]time.Time

	// Frecency : A coined word of "frequently" + "recency"
	Frecency struct {
		Word  string `json:"word"`
		Score int    `json:"score"`
	}
	// History : List of Frecency sorted by Frecency.Score
	History []Frecency
)

// logWord extract search word from logfile
func logWord(logfile string) (historyMap, error) {
	history := make(historyMap, 100)
	fp, err := os.Open(logfile)
	if err != nil {
		return history, err
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	for {
		var (
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
		// slog text形式: msg=search かつ cache=PUSH or cache=GET の行のみ処理
		if !strings.Contains(lines, `msg=search`) {
			continue
		}
		if !strings.Contains(lines, `cache=PUSH`) && !strings.Contains(lines, `cache=GET`) {
			continue
		}
		// 検索エラーのない文字列だけfrecencyに追加する
		sw, ew, err := api.QueryParser(ExtractKeyword(lines))
		if err != nil {
			continue // Ignore QueryParser() Error
		}

		word := Normalize(sw, ew)
		event, err = ExtractDatetime(lines)
		if err != nil {
			continue // Ignore time.Parse() Error
		}
		history[word] = append(history[word], event)
	}
	return history, err
}

// ExtractDatetime extract search datetime from a line of slog text format
// Log: time=2024-01-15T10:30:00Z level=INFO msg=search ...
func ExtractDatetime(s string) (time.Time, error) {
	re := regexp.MustCompile(`time=(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})`)
	layout := "2006-01-02T15:04:05"
	m := re.FindStringSubmatch(s)
	if m == nil {
		return time.Time{}, &time.ParseError{}
	}
	return time.Parse(layout, m[1])
}

// ExtractKeyword extract search keyword from a line of slog text format
// Log: ... query="usr pac" or query=usr
func ExtractKeyword(s string) string {
	re := regexp.MustCompile(`query=("([^"]+)"|(\S+))`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return s
	}
	if m[2] != "" {
		return m[2] // quoted value
	}
	return m[3] // unquoted value
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

// ScoreSum : 履歴マップの検索日時リストからスコア合計を算出する
func ScoreSum(tl []time.Time) (score int) {
	for _, t := range tl {
		score += Scoring(t)
	}
	return
}

// RankByScore : 履歴から頻出度リストを生成する
func (h historyMap) RankByScore() History {
	var i int
	l := make(History, len(h))
	for k, v := range h {
		l[i] = Frecency{k, ScoreSum(v)}
		i++
	}
	sort.Sort(l)
	return l
}

func (his History) Len() int           { return len(his) }
func (his History) Less(i, j int) bool { return his[i].Score > his[j].Score }
func (his History) Swap(i, j int)      { his[i], his[j] = his[j], his[i] }

// Datalist throw map of searched words sorted by score
func Datalist(f string) (History, error) {
	historyMap, err := logWord(f)
	history := historyMap.RankByScore()
	return history, err
}

// Filter returns gt (greter than) and lt (less than) score of history
func (his History) Filter(gt, lt int) (filtered History) {
	for _, h := range his {
		if h.Score > gt && h.Score < lt {
			filtered = append(filtered, h)
		}
	}
	return
}
