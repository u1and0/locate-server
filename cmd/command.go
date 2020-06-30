package locater

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	pipeline "github.com/mattn/go-pipeline"
)

// PathMap is pairs of fullpath:dirpath
type PathMap struct {
	File      string
	Dir       string
	Highlight string
}

// Stats : locate検索の統計情報
type Stats struct {
	LastUpdateTime string  // 最後のDBアップデート時刻
	SearchTime     float64 // 検索にかかった時間
	ResultNum      int     // 検索結果数
	Items          int     // 検索対象のすべてのファイル数
}

// LocateStats : Result of `locate -S`
func LocateStats(path string) ([]byte, error) {
	opt := []string{"-S"}
	if path != "" {
		opt = append(opt, "-d", path)
	}
	b, err := exec.Command("locate", opt...).Output()
	return b, err
}

// LocateStatsSum : locateされるファイル数をDB情報から合計する
func LocateStatsSum(b []byte) int {
	var sum, ni int
	for i, w := range strings.Split(string(b), "\n") { // 改行区切り => 221,453 ファイル
		if i%5 == 2 {
			ns := strings.Fields(w)[0]           // => 221,453
			ns = strings.ReplaceAll(ns, ",", "") // => 221453
			ni, _ = strconv.Atoi(ns)
			sum += ni
		}
	}
	return sum
}

// sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words []string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`)
		/* Replace only a word
		全て変える
		re.ReplaceAll(s, "<span style=\"background-color:#FFCC00;\">$1</span>")
		は削除
		*/
		color := "style=\"background-color:#FFCC00;\">"
		found := re.FindString(s)
		if found != "" {
			s = strings.Replace(s,
				found,
				"<span "+color+found+"</span>",
				1)
			// [BUG] キーワード順にハイライトされない
		}
	}
	return s
}

// CmdGen : locate 検索語 | grep -v 除外語 | grep -v 除外語...を発行する
func (l *Locater) CmdGen() [][]string {
	// -i: Ignore case distinctions when matching patterns.
	locate := []string{"locate",
		"--ignore-case",
		"--quiet", // report no error messages about reading databases
	}
	if l.Dbpath != "" {
		// -d: Replace the default database with DBPATH.
		locate = append(locate, "-d", l.Dbpath)
	}

	// Include PATTERNs
	locate = append(locate, "--regex", strings.Join(l.SearchWords, ".*"))
	// -> hoge.*my.*name

	// Exclude PATTERNs
	exec := [][]string{locate}
	for _, ex := range l.ExcludeWords {
		exec = append(exec, []string{"grep", "-ivE", ex})
	}
	return exec
}

// Cmd : locate検索し、
// 結果をPathMapのスライス(最大l.Limit件(limit = default 1000))にして返す
// 更に検索結果数、あれば検索時のエラーを返す
func (l *Locater) Cmd() ([]PathMap, int, error) {
	out, err := pipeline.Output(l.CmdGen()...)
	outslice := strings.Split(string(out), "\n")
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n

	results := make([]PathMap, 0, l.Limit)
	/* Why not array but slice?
	検索結果の数だけ要素を持ったスライスを返したい
	検索結果がなければ0要素のスライスを返したい
	そのため、要素数の決まった配列を使えない
	> 後で空の要素は削除して結果に表示しないようにしないといけない
	最大の要素数はlimit(デフォルト1000件)になるように表示する
	*/
	for i, file := range outslice {
		// l.Limit件までresultsとして返す
		if i >= l.Limit {
			break
		}

		/* 親ディレクトリ */
		dir := filepath.Dir(file)

		/* オプションによる結果の変換
		1. UNIXドライブパスを取り除いて
		2. Windowsパスセパレータ(\)に変換して
		3. Windows or UNIX ルートドライブパスを取り付ける
		順番は大事
		*/
		if l.Trim != "" { // Trim drive path
			file = strings.TrimPrefix(file, l.Trim)
			dir = strings.TrimPrefix(dir, l.Trim)
		}
		if l.PathSplitWin { // Transfer separator
			file = strings.ReplaceAll(file, "/", "\\")
			dir = strings.ReplaceAll(dir, "/", "\\")
		}
		if l.Root != "" { // Insert drive path
			file = l.Root + file
			dir = l.Root + dir
		}

		/* 検索キーワードをハイライト */
		highlight := highlightString(file, l.SearchWords)

		/* 最終的な表示結果をresultsに代入
		見つかった結果の分だけsliceを拡張する
		*/
		results = append(results, PathMap{file, dir, highlight})
	}

	// Max 1000 result & number of all result
	return results, len(outslice), err
}
