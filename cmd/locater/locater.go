package locater

import (
	"strconv"
	"strings"

	pipeline "github.com/mattn/go-pipeline"
	api "github.com/u1and0/locate-server/cmd/api"
)

type (
	// Locater : queryから読み取った検索ワードと無視するワード
	Locater struct {
		// Input
		Args      `json:"args"`
		api.Query `json:"query"`
		// Extract
		SearchWords  []string `json:"searchWords"`  // 検索キーワード
		ExcludeWords []string `json:"excludeWords"` // 検索から取り除くキーワード
		// Output
		Paths `json:"paths"`
		Stats `json:"stats"`
		Error string `json:"error"`
	}

	// Args is command line option
	Args struct {
		Dbpath       string `json:"dbpath"`       // 検索対象DBパス /path/to/database:/path/to/another
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
	}
)

// Locate excute locate (or gocate) command
// split from Locater.Cmd()
func (l *Locater) Locate() (Paths, error) {
	out, err := pipeline.Output(l.CmdGen()...)
	outslice := strings.Split(string(out), "\n")
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n
	return outslice, err
}

// CmdGen : shell実行用パイプラインコマンドを発行する
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

	// Limit option
	if l.Query.Limit > 0 {
		pipeline = append(pipeline, []string{"head", "-n", strconv.Itoa(l.Query.Limit)})
	}

	if l.Args.Debug {
		log.Debugf("Execute command %v", pipeline)
	}
	return  // => locate ... | grep -ivE ... | head -n ...
}
