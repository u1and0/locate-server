package locater

import (
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
		Paths `json:"paths"`
		Stats `json:"stats"`
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
