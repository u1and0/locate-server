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
	ResultNum      uint64  // 検索結果数
	Items          string  // 検索対象のすべてのファイル数
}

// LocateStats : Result of `locate -S`
func LocateStats(path string) ([]byte, error) {
	opt := []string{"-S"}
	if path != "" {
		opt = append(opt, "--database", path)
	}
	b, err := exec.Command("locate", opt...).Output()
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
	case n >= 1e6:
		s = strconv.FormatUint(n/1e6, 10) + "百万"
	case n >= 1e4:
		s = strconv.FormatUint(n/1e4, 10) + "万"
	case n >= 1e3:
		s = strconv.FormatUint(n/1e3, 10) + "千"
	default:
		s = strconv.FormatUint(n, 10)
	}
	return
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

// CmdGen : shell実行用パイプラインコマンドを発行する
//
// Process = 1のとき
// locate 検索語 | grep -v 除外語 | grep -v 除外語...
//
// Process = 1以外のとき
// マルチプロセスlocateを発行する
// echo DBPATH | sed -e 's/:/\n/g'| xargs -P0 -I@ locate 検索語 | grep -v 除外語 | grep -v 除外語...
func (l *Locater) CmdGen() (pipeline [][]string) {
	locate := []string{
		"locate",
		"--ignore-case", // Ignore case distinctions when matching patterns.
		"--quiet",       // Report no error messages about reading databases
	}

	// Include PATTERNs
	// -> locate --ignore-case --quiet --regex hoge.*my.*name
	locate = append(locate, "--regex", strings.Join(l.SearchWords, ".*"))

	if l.Process != 1 { // Multi processing search
		echo := []string{"echo", l.Dbpath}
		sed := []string{"sed", "-e", "s/:/\\n/g"}
		// xargs -P 2 -I@
		xargs := []string{"xargs", "-P", strconv.Itoa(l.Process), "-I@"}
		// xargs -P 2 -I@ locate -iq --regex hoge.*foo
		xargs = append(xargs, locate...)
		// xargs -P 2 -I@ locate -iq --regex hoge.*foo --database @
		xargs = append(xargs, "--database @")
		// echo /path/to/some.db:/path/to/another.db |
		//		sed -e 's/:/\n/g' |
		//		xargs -P 2 -I@ locate -iq --regex hoge.*foo --database @
		pipeline = append(pipeline, echo, sed, xargs)
	} else { // Single processing search
		if l.Dbpath != "" { // Replace the default database to Dbpath
			locate = append(locate, "--database", l.Dbpath)
		}
		pipeline = append(pipeline, locate)
	}

	// Exclude PATTERNs
	for _, ex := range l.ExcludeWords {
		// COMMAND | grep -ivE EXCLUDE1 | grep -ivE EXCLUDE2
		pipeline = append(pipeline, []string{"grep", "-ivE", ex})
	}
	return
}

// Cmd : locate検索し、
// 結果をPathMapのスライス(最大l.Limit件(limit = default 1000))にして返す
// 更に検索結果数、あれば検索時のエラーを返す
func (l *Locater) Cmd() ([]PathMap, uint64, error) {
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
	return results, uint64(len(outslice)), err
}
