package locater

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
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

// plocate not implement -S option
// // LocateStats : Result of `locate -S`
// func LocateStats(s string) ([]byte, error) {
// 	dbs, err := filepath.Glob(s + "/*.db")
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	d := strings.Join(dbs, ":")
// 	b, err := exec.Command("locate", "-Sd", d).Output()
// 	// => locate -Sd /var/lib/mlocate/db1.db:/var/lib/mlocate/db2.db:...
// 	if err != nil {
// 		return b, err
// 	}
// 	return b, err
// }

// // LocateStatsSum : locateされるファイル数をDB情報から合計する
// func LocateStatsSum(b []byte) (int64, error) {
// 	var (
// 		sum, ni int64
// 		ns      string
// 		err     error
// 	)
// 	for i, w := range strings.Split(string(b), "\n") { // 改行区切り => 221,453 ファイル
// 		if i%5 == 2 {
// 			ns = strings.Fields(w)[0]              // => 221,453
// 			ns = strings.ReplaceAll(ns, ",", "")   // => 221453
// 			ni, err = strconv.ParseInt(ns, 10, 64) // as int64
// 			if err != nil {
// 				return sum, err
// 			}
// 			sum += ni
// 		}
// 	}
// 	return sum, err
// }

// LocateStats : count all file num on db
func LocateStats(s string) ([]byte, error) {
	dbs, err := filepath.Glob(s + "/*.db")
	if err != nil {
		return []byte{}, err
	}
	d := strings.Join(dbs, ":")
	b, err := exec.Command("locate", "-d", d, "-c", "--regex", ".").CombinedOutput()
	// => locate -d /var/lib/plocate/db1.db:/var/lib/plocate/db2.db:... -c --regex "."
	if err != nil {
		return b, err
	}
	return b, err
}

func LocateStatsSum(b []byte) (int64, error) {
	s := strings.Fields(string(b))[0]
	return strconv.ParseInt(s, 10, 64)

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
