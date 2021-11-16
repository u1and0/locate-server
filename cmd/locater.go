package locater

import (
	"path/filepath"
	"sort"
	"strings"

	pipeline "github.com/mattn/go-pipeline"
)

type (
	// Locater : queryから読み取った検索ワードと無視するワード
	Locater struct {
		SearchWords  []string `json:"searchWords"`  // 検索キーワード
		ExcludeWords []string `json:"excludeWords"` // 検索から取り除くキーワード
		Args         `json:"args"`
		// -- Result struct
		Paths     `json:"paths"`
		Highlight `json:"highlight"`
		Stats     `json:"stats"`
	}

	// Args is command line option
	Args struct {
		Dbpath       string `json:"dbpath"`       // 検索対象DBパス /path/to/database:/path/to/another
		Limit        int    `json:"limit"`        // 検索結果HTML表示制限数
		PathSplitWin bool   `json:"pathSplitWin"` // TrueでWindowsパスセパレータを使用する
		Root         string `json:"root"`         // 追加するドライブパス名
		Trim         string `json:"trim"`         // 削除するドライブパス名
		Debug        bool   `json:"debug"`        // Debugフラグ
	}

	// Paths locate command result
	Paths []string
	// Highlight locate command result with HTML colored background
	Highlight []string

	// Stats : locate検索の統計情報
	Stats struct {
		LastUpdateTime string  `json:"lastUpdateTime"` // 最後のDBアップデート時刻
		SearchTime     float64 `json:"searchTime"`     // 検索にかかった時間
		Items          string  `json:"items"`          // 検索対象のすべてのファイル数
		Response       int     `json:"response"`       // httpレスポンス　成功で200
	}
)

// Normalize : SearchWordsとExcludeWordsを合わせる
// SearchWordsは小文字にする
// ExcludeWordsは小文字にした上で
// ソートして、頭に-をつける
func (l *Locater) Normalize() string {
	se := l.SearchWords
	ex := l.ExcludeWords

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

// Locate excute locate (or gocate) command
// split from Locater.Cmd()
func (l *Locater) Locate() (Paths, error) {
	out, err := pipeline.Output(l.CmdGen()...)
	if l.Debug {
		log.Debugf("gocate result %v", out)
	}
	outslice := strings.Split(string(out), "\n")
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n
	return outslice, err
}

// Convert modify Paths by command line Args
// split from Locater.Cmd()
func (l *Locater) Convert(p Paths) Paths {
	q := make(Paths, len(p))
	for i, file := range p {
		/* オプションによる結果の変換
		1. UNIXドライブパスを取り除いて
		2. Windowsパスセパレータ(\)に変換して
		3. Windows or UNIX ルートドライブパスを取り付ける
		順番は大事
		*/
		if l.Trim != "" { // Trim drive path
			file = strings.TrimPrefix(file, l.Trim)
		}
		if l.PathSplitWin { // Transfer separator
			file = strings.ReplaceAll(file, "/", "\\")
		}
		if l.Root != "" { // Insert drive path
			file = l.Root + file
		}
		q[i] = file
	}
	return q
}

// CmdGen : shell実行用パイプラインコマンドを発行する
//
// Process = 1のとき
// locate 検索語 | grep -v 除外語 | grep -v 除外語...
//
// Process = 1以外のとき
// マルチプロセスlocateを発行する
// echo $LOCATE_PATH | tr :, \n | xargs -P0 -I@ locate 検索語 | grep -v 除外語 | grep -v 除外語...
func (l *Locater) CmdGen() (pipeline [][]string) {
	locate := []string{
		"gocate",               // locate command path
		"--database", l.Dbpath, //Add database option
		"--",            // Inject locate option
		"--ignore-case", // Ignore case distinctions when matching patterns.
		"--quiet",       // Report no error messages about reading databases
		"--existing",    // Print only entries that refer to files existing at the time locate is run.
		"--nofollow",    // When  checking  whether files exist do not follow trailing symbolic links.
	}
	// -> gocate --database -- --ignore-case --quiet --regex hoge.*my.*name

	// Include PATTERNs
	// -> locate --ignore-case --quiet --regex hoge.*my.*name
	locate = append(locate, "--regex", strings.Join(l.SearchWords, ".*"))

	pipeline = append(pipeline, locate)

	// Exclude PATTERNs
	for _, ex := range l.ExcludeWords {
		// COMMAND | grep -ivE EXCLUDE1 | grep -ivE EXCLUDE2
		pipeline = append(pipeline, []string{"grep", "-ivE", ex})
	}
	if l.Debug {
		log.Debugf("Execute command %v", pipeline)
	}
	return
}

// Cmd : locate検索し、
// 結果をPathMapのスライス(最大l.Limit件(limit = default 1000))にして返す
// 更に検索結果数、あれば検索時のエラーを返す
func (l *Locater) Cmd() ([]PathMap, uint64, error) {
	results := make([]PathMap, 0, l.Limit)
	var resultsNum uint64

	out, err := pipeline.Output(l.CmdGen()...)
	if err != nil {
		return results, resultsNum, err
	}
	outslice := strings.Split(string(out), "\n")
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n
	resultsNum = uint64(len(outslice))

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
		highlight := HighlightString(file, l.SearchWords)

		/* 最終的な表示結果をresultsに代入
		見つかった結果の分だけsliceを拡張する
		*/
		results = append(results, PathMap{file, dir, highlight})
	}

	// Max 1000 result & number of all result
	return results, resultsNum, err
}
