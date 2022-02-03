package locater

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("locater")

// PathMap is pairs of fullpath:dirpath
type PathMap struct {
	File      string
	Dir       string
	Highlight string
}

// DBLastUpdateTime returns date time string for directory update time
func DBLastUpdateTime(db string) string {
	filestat, err := os.Stat(db)
	if err != nil {
		log.Error(err)
	}
	layout := "2006-01-02 15:05"
	return filestat.ModTime().Format(layout)
}

// LocateStats : Sum db size
func LocateStats(s string) (int64, error) {
	var sum int64
	dbs, err := filepath.Glob(s + "/*.db")
	if err != nil {
		return sum, err
	}
	for _, d := range dbs {
		file, err := os.Open(d)
		defer file.Close()
		i, err := file.Stat()
		s := i.Size()
		if err != nil {
			return s, err
		}
		sum += s
	}
	return sum, err
}

// Ambiguous : 数値を切り捨て、おおよその数字をstring型にして返す
// 684,345(int) => 680,000+(string)
func Ambiguous(n int64) (s string) {
	switch {
	case n >= 1e9:
		s = humanize.Comma(dropDigit(n, 1_000_000_000)) + "+"
	case n >= 1e6:
		s = humanize.Comma(dropDigit(n, 1_000_000)) + "+"
	case n >= 1e3:
		s = humanize.Comma(dropDigit(n, 1_000)) + "+"
	default:
		s = humanize.Comma(n)
	}
	return
}

func dropDigit(n, m int64) int64 {
	return n / m * m
}

// Normalize : SearchWordsとExcludeWordsを合わせる
// SearchWordsは小文字にする
// ExcludeWordsは小文字にした上で
// ソートして、頭に-をつける
func Normalize(se, ex []string) string {
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
