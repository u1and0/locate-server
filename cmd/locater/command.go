package locater

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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

// LocateStats : Result of `locate -S`
func LocateStats(s string) ([]byte, error) {
	dbs, err := filepath.Glob(s + "/*.db")
	if err != nil {
		return []byte{}, err
	}
	d := strings.Join(dbs, ":")
	b, err := exec.Command("locate", "-Sd", d).Output()
	// => locate -Sd /var/lib/mlocate/db1.db:/var/lib/mlocate/db2.db:...
	if err != nil {
		return b, err
	}
	return b, err
}

// LocateStatsSum : locateされるファイル数をDB情報から合計する
func LocateStatsSum(b []byte) (uint64, error) {
	var (
		sum, ni uint64
		err     error
	)
	for i, w := range strings.Split(string(b), "\n") { // 改行区切り => 221,453 ファイル
		if i%5 == 2 {
			ns := strings.Fields(w)[0]             // => 221,453
			ns = strings.ReplaceAll(ns, ",", "")   // => 221453
			ni, err = strconv.ParseUint(ns, 10, 0) // as uint64
			if err != nil {
				return sum, err
			}
			sum += ni
		}
	}
	return sum, err
}

// Ambiguous : 数値を切り捨て、おおよその数字をstring型にして返す
func Ambiguous(n uint64) (s string) {
	switch {
	case n >= 1e8:
		s = strconv.FormatUint(n/1e8, 10) + "億"
	case n >= 1e4:
		s = strconv.FormatUint(n/1e4, 10) + "万"
	case n >= 1e3:
		s = strconv.FormatUint(n/1e3, 10) + "千"
	default:
		s = strconv.FormatUint(n, 10)
	}
	return
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
