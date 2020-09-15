package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	cmd "locate-server/cmd"

	"github.com/op/go-logging"
)

const (
	// VERSION : version
	VERSION = "2.0.0r"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// LOCATEDIR : locateのデータベースやログファイルを置く場所
	// /var/lib/mlocate以下すべてを検索対象とするときのLOCATE_PATHの指定方法
	// LOCATE_PATH=$(paste -sd: <(find /var/lib/mlocate -name '*.db'))
	LOCATEDIR = "/var/lib/mlocate"
	// DEFAULTDB : locateがデフォルトで検索するdbpath
	DEFAULTDB = "/var/lib/mlocate/mlocate.db"
)

var (
	showVersion  bool
	receiveValue string
	err          error
	limit        int
	// dbpathオプションがなければ$LOCATE_PATHを参照する
	// $LOCATE_PATHも空のときはデフォルト値"/var/lib/mlocate/mlocate.db"を使用する
	dbpath       = "/var/lib/mlocate/mlocate.db"
	pathSplitWin bool
	root         string
	trim         string
	cache        cmd.CacheMap
	getpushLog   string
	locateS      []byte
	stats        cmd.Stats
	process      int
	debug        bool
)

var log = logging.MustGetLogger("main")

func main() {
	flag.StringVar(&dbpath, "d", DEFAULTDB,
		"Path of locate database file (ex: /path/something.db:/path/another.db)")

	flag.IntVar(&limit, "l", 1000, "Maximum limit for results")
	flag.BoolVar(&pathSplitWin, "s", false, "OS path split windows backslash")
	flag.StringVar(&root, "r", "", "DB insert prefix for directory path")
	flag.StringVar(&trim, "t", "", "DB trim prefix for directory path")
	flag.IntVar(&process, "P", 1, "Search in multi process by `xargs -P`")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	// Version info
	// フラグ解析以降はここより後に書く
	if showVersion {
		fmt.Println("version:", VERSION)
		return // versionを表示して終了
	}

	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer logfile.Close()
	setLogger(logfile) // log.XXX()を使うものはここより後に書く
	if err != nil {
		log.Panicf("Cannot open logfile %v", err)
	}

	// DB path flag parse
	// dbpathがデフォルトであり かつ $LOCATE_PATHが指定されていれば(=空でなければ)
	if d := os.Getenv("LOCATE_PATH"); dbpath == DEFAULTDB && d != "" {
		dbpath = d // dbpathを上書きする
		if err := os.Setenv("LOCATE_PATH", ""); err != nil {
			log.Panicf("Cannot set env variable %s", err)
		}
	}
	log.Infof("Set dbpath: %s", dbpath)

	// Directory check
	if _, err := os.Stat(LOCATEDIR); os.IsNotExist(err) {
		log.Panic(err) // /var/lib/mlocateがなければ終了
	}

	// Command check
	if _, err := exec.LookPath("locate"); err != nil {
		log.Panic(err) // locateコマンドがなければ終了
	}

	// Initialize cache
	// nil map assignment errorを発生させないために必要
	cache = cmd.CacheMap{}
	// cacheを廃棄するかの判断に必要
	// lstatが変わった=mlocate.dbの内容が更新されたのでcacheを新しくする
	locateS, err = cmd.LocateStats(dbpath)
	if err != nil {
		log.Error(err)
	}
	// 検索対象ファイル数の合計値を算出
	var n uint64
	n, err = cmd.LocateStatsSum(locateS)
	if err != nil {
		log.Error(err)
	}
	stats.Items = cmd.Ambiguous(n)
	if err != nil {
		log.Error(err)
	}

	// HTTP pages
	http.HandleFunc("/", showInit)
	http.HandleFunc("/searching", addResult)
	http.HandleFunc("/status", locateStatusPage)
	http.ListenAndServe(":8080", nil)
}

// setLogger is printing out log message to STDOUT and LOGFILE
func setLogger(f *os.File) {
	var format = logging.MustStringFormatter(
		`%{color}[%{level:.6s}] ▶ %{time:2006/01/02 15:04:05.000} %{shortfile} %{message} %{color:reset}`,
	)
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(f, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)
}

// html デフォルトの説明文
func htmlClause(s string) string {
	logs, err := cmd.LogWord()
	if err != nil {
		log.Error(err)
	}
	sw := Datalist(logs)
	return fmt.Sprintf(`<html>
					<head><title>Locate Server %s</title></head>
					<body>
						<form method="get" action="/searching">
							<input type="text" name="query" value="%s" size="50" list="searchedWords">
							<datalist id="searchedWords">
							%s
							</datalist>
							<input type="submit" name="submit" value="検索">
							<a href=https://github.com/u1and0/locate-server/blob/master/README.md>Help</a>
						</form>
						<small>
							 * 検索文字列は2文字以上を指定してください。<br>
							 * 英字の大文字/小文字は無視します。<br>
							 * << マーククリックでフォルダが開きます。<br>
							 * スペース区切りで複数入力できます。(AND検索)<br>
							 * 半角カッコでくくって | で区切ると | で区切られる前後で検索します。(OR検索)<br>
							 例: "電(気|機)工業" => "電気工業"と"電機工業"を検索します。<br>
							 * 単語の頭に半角ハイフン"-"をつけるとその単語を含まないファイルを検索します。(NOT検索)<br>
							 例: "電気 -工 業"=>"電気"と"業"を含み"工"を含まないファイルを検索します。
							 </small>`, s, s, sw)
}

// Datalist convert []string to <datalist> string
func Datalist(slice []string) string {
	var list []string
	for _, l := range slice {
		list = append(list, fmt.Sprintf(`<option value="%s"></option>`, l))
	}
	return strings.Join(list, "")
}

// Top page
func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause(""))
}

// `locate -S` page
func locateStatusPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html>
					<head><title>Locate DB Status</title></head>
					<body>
						<pre>%s</pre>
					</body>
					</html>`,
		func() (s interface{}) {
			if l, err := cmd.LocateStats(dbpath); err == nil {
				s = l
			} else {
				s = err.Error()
			}
			return
		}(),
	)
}

// locate検索し、結果をhtmlに書き込む
func addResult(w http.ResponseWriter, r *http.Request) {
	var (
		results []cmd.PathMap
	)
	// 検索コンフィグ構造体
	loc := cmd.Locater{
		Limit:        limit,        // 検索件数上限
		Dbpath:       dbpath,       // 検索対象パス
		PathSplitWin: pathSplitWin, // path separatorを\にする
		Root:         root,         // Path prefix insert
		Trim:         trim,         // Path prefix trim
		Process:      process,      // xargsによるマルチプロセス数
		Debug:        debug,        //Debugフラグ
	}
	// Modify query
	receiveValue = r.FormValue("query")

	if loc.SearchWords, loc.ExcludeWords, err =
		cmd.QueryParser(receiveValue); err != nil { // 検索文字チェックERROR
		log.Errorf("%s [ %-50s ]", err, receiveValue)
		fmt.Fprint(w, htmlClause(receiveValue))
		fmt.Fprintf(w, `<h4>
							%s
						</h4>
					</body>
					</html>`, err)
	} else { // 検索文字数チェックOK
		/* LocateStats()の結果が前と異なっていたら
		locateS更新
		cacheを初期化 */
		if l, err := cmd.LocateStats(dbpath); string(l) != string(locateS) { // DB更新されていたら
			if err != nil {
				log.Error(err)
			}
			locateS = l // 保持するDB情報の更新
			var n uint64
			n, err = cmd.LocateStatsSum(l) // 検索ファイル数の更新
			stats.Items = cmd.Ambiguous(n)
			if err != nil {
				log.Error(err)
			}
			cache = cmd.CacheMap{} // キャッシュリセット
		}

		// Searching
		st := time.Now()
		results, stats.ResultNum, getpushLog, err = loc.ResultsCache(&cache)
		/* cache は&cacheによりdeep copyされてResultsCache()内で
		直接書き換えられるので、returnされない*/
		stats.SearchTime = float64((time.Since(st)).Nanoseconds()) / float64(time.Millisecond)

		if err != nil {
			log.Errorf("%s [ %-50s ]", err, receiveValue)
		}
		log.Noticef("%8dfiles %3.3fmsec %s [ %-50s ]",
			stats.ResultNum, stats.SearchTime, getpushLog, receiveValue)
		/* normalizedWordではなく、あえてreceiveValueを
		表示して生の検索文字列を記録したい*/

		// Update time
		filestat, err := os.Stat(LOCATEDIR)
		if err != nil {
			log.Error(err)
		}
		layout := "2006-01-02 15:05"
		stats.LastUpdateTime = filestat.ModTime().Format(layout)

		// Search result page
		fmt.Fprint(w, htmlClause(receiveValue))

		// Google検索の表示例
		// 約 8,010,000 件 （0.50 秒）
		fmt.Fprintf(w,
			`<h4>
			 <a href=/status>DB</a> last update: %s<br>
			 ヒット数: %d件中、最大%d件を表示<br>
			 %.3fmsec で約%s件を検索しました。<br>
			</h4>`,
			stats.LastUpdateTime,
			stats.ResultNum,
			loc.Limit,
			stats.SearchTime,
			stats.Items)

		// 検索結果を行列表示
		fmt.Fprintln(w, `<table>
						  <tr>`)
		for _, e := range results {
			fmt.Fprintf(w, `<tr>
				<td>
					<a href="file://%s">%s</a>
					<a href="file://%s" title="<< クリックでフォルダに移動"><<</a>
				</td>
			</tr>`, e.File, e.Highlight, e.Dir)
		}

		fmt.Fprintln(w, `</table>
					  </body>
					  </html>`)
	}
}
