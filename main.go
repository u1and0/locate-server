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
	VERSION = "2.3.2"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// LOCATEDIR : locateのデータベースやログファイルを置く場所
	// /var/lib/mlocate以下すべてを検索対象とするときのLOCATE_PATHの指定方法
	// LOCATE_PATH=$(paste -sd: <(find /var/lib/mlocate -name '*.db'))
	LOCATEDIR = "/var/lib/mlocate"
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
	debug        bool
)

var log = logging.MustGetLogger("main")

func main() {
	flag.StringVar(&dbpath, "d", LOCATEDIR, "Path of locate database directory")
	flag.IntVar(&limit, "l", 1000, "Maximum limit for results")
	flag.BoolVar(&pathSplitWin, "s", false, "OS path split windows backslash")
	flag.StringVar(&root, "r", "", "DB insert prefix for directory path")
	flag.StringVar(&trim, "t", "", "DB trim prefix for directory path")
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
	log.Infof("Set dbpath: %s", dbpath)

	// Directory check
	if _, err := os.Stat(LOCATEDIR); os.IsNotExist(err) {
		log.Panic(err) // /var/lib/mlocateがなければ終了
	}

	// Command check
	if _, err := exec.LookPath("gocate"); err != nil {
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
	n, err = cmd.LocateStatsSum(locateS) // 検索ファイル数の初期値
	if err != nil {
		log.Error(err)
	}
	stats.Items = cmd.Ambiguous(n)
	stats.LastUpdateTime = DBLastUpdateTime()
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
		`%{color}[%{level:.6s}] ▶ %{time:2006-01-02 15:04:05} %{shortfile} %{message} %{color:reset}`,
	)
	backend1 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2 := logging.NewLogBackend(f, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	logging.SetBackend(backend1Formatter, backend2Formatter)
}

// html デフォルトの説明文
func htmlClause(title string) string {
	// Get searched word from log file
	historymap, err := cmd.LogWord(LOGFILE)
	if err != nil {
		log.Error(err)
	}
	wordList := historymap.RankByScore()
	if debug {
		log.Debugf("Frecency list: %v", wordList)
	}
	explain := SurroundTag(func() string {
		ss := []string{
			`検索ワードを指定して検索を押すかEnterキーを押すと共有フォルダ内のファイルを高速に検索します。`,
			`対象文字列は2文字以上の文字列を指定してください。`,
			`英字 大文字/小文字は無視します。`,
			`全角/半角スペースで区切ると0文字以上の正規表現(.*)に変換して検索されます。(AND検索)`,
			`"(aaa|bbb)"のグループ化表現が使えます。(OR検索)` +
				SurroundTag(
					SurroundTag(
						fmt.Sprintf(`例: %s => %s並びに%sを検索します。`,
							SurroundTag(`golang (pdf|txt)`, "strong"),
							SurroundTag("golang及びpdf", "strong"),
							SurroundTag("golang及びtxt", "strong"),
						),
						"li",
					),
					"ul",
				),
			`[a-zA-Z0-9]の正規表現が使えます。` +
				SurroundTag(
					SurroundTag(
						fmt.Sprintf(`例: %s =>%s並びに%s を検索します。`,
							SurroundTag("file[xy] txt", "strong"),
							SurroundTag("filex及びtxt", "strong"),
							SurroundTag("filey及びtxt", "strong"),
						),
						"li",
					)+
						SurroundTag(
							fmt.Sprintf(`例: %s =>%s、%s並びに%s を検索します。`,
								SurroundTag("file[x-z] txt", "strong"),
								SurroundTag("filex及びtxt", "strong"),
								SurroundTag("filey及びtxt", "strong"),
								SurroundTag("filez及びtxt", "strong"),
							),
							"li",
						)+
						SurroundTag(
							fmt.Sprintf(`例: %s  => %s, %s, %s, %sを検索します。`,
								SurroundTag("201[6-9]S", "strong"),
								SurroundTag("2016S", "strong"),
								SurroundTag("2017S", "strong"),
								SurroundTag("2018S", "strong"),
								SurroundTag("2019S", "strong"),
							),
							"li",
						),
					"ul",
				),
			`0文字か1文字の正規表現"?"が使えます。` +
				SurroundTag(
					SurroundTag(
						fmt.Sprintf(`例: %s => %sと %sを検索します。`,
							SurroundTag("jpe?g", "strong"),
							SurroundTag("jpeg", "strong"),
							SurroundTag("jpg", "strong"),
						),
						"li",
					),
					"ul",
				),
			`単語の頭に半角ハイフン"-"をつけるとその単語を含まないファイルを検索します。(NOT検索)` +
				SurroundTag(
					SurroundTag(
						fmt.Sprintf(`例: %s=>%sと%sを含み%sを含まないファイルを検索します。`,
							SurroundTag("gobook txt -doc", "strong"),
							SurroundTag("gobook", "strong"),
							SurroundTag("txt", "strong"),
							SurroundTag("doc", "strong"),
						),
						"li",
					),
					"ul",
				),
			`AND検索は順序を守って検索をかけますが、NOT検索は順序は問わずに除外します。` +
				SurroundTag(
					SurroundTag(
						fmt.Sprintf(`例: %s と%s は異なる検索結果ですが、 %s と%sは同じ検索結果になります。`,
							SurroundTag("gobook txt -doc", "strong"),
							SurroundTag("txt gobook -doc", "strong"),
							SurroundTag("gobook txt -doc", "strong"),
							SurroundTag("gobook -doc txt", "strong"),
						),
						"li",
					),
					"ul",
				),
			fmt.Sprintf(`ファイル拡張子を指定するときは、文字列の最後を表す%s記号を行末につけます。`, SurroundTag("$", "strong")) +
				SurroundTag(
					SurroundTag(
						fmt.Sprintf(`例: %s =>%sを含み、%sが行末につくファイルを検索します。`,
							SurroundTag("gobook pdf$", "strong"),
							SurroundTag("gobook", "strong"),
							SurroundTag("pdf", "strong"),
						),
						"li",
					),
					"ul",
				),
		}
		for i, s := range ss {
			ss[i] = SurroundTag(s, "li")
		}
		return strings.Join(ss, "")
	}(), "ul")
	return fmt.Sprintf(`<html>
					<head><title>Locate Server %s</title></head>
					<body>
						<form method="get" action="/searching">
							<!-- 検索窓 -->
							<input type="text" name="query" value="%s" size="50" list="searched-words" >

							<!-- 検索履歴 Frecency リスト -->
 							<datalist id="searched-words"> %s </datalist>

							<!-- 検索ボタン -->
							<input type="submit" name="submit" value="検索">
							<a href=https://github.com/u1and0/locate-server/blob/master/README.md>Help</a>
						</form>

						<!-- 折りたたみ展開ボタン -->
						<div onclick="obj=document.getElementById('hidden-explain').style; obj.display=(obj.display=='none')?'block':'none';">
						<a style="cursor:pointer;">▼ 検索ヘルプを表示</a>
						</div>
						<!--// 折りたたみ展開ボタン -->

						<!-- ここから先を折りたたむ -->
						<div id="hidden-explain" style="display:none;clear:both;">
						<!-- 検索ヘルプ -->
						<small> %s </small>
						</div>
						<!-- 折りたたみここまで -->

						<h4>
							<a href=/status>DB</a> last update: %s<br>
						`,
		title,
		title,
		wordList.Datalist(),
		explain,
		stats.LastUpdateTime,
	)
}

// SurroundTag surrounds some word `s` for any html tag `tag`
func SurroundTag(s, tag string) string {
	return fmt.Sprintf("<%s>%s</%s>", tag, s, tag)
}

// DBLastUpdateTime returns date time string for directory update time
func DBLastUpdateTime() string {
	filestat, err := os.Stat(LOCATEDIR)
	if err != nil {
		log.Error(err)
	}
	layout := "2006-01-02 15:05"
	return filestat.ModTime().Format(layout)
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
		Debug:        debug,        //Debugフラグ
	}
	// Modify query
	receiveValue = r.FormValue("query")

	if loc.SearchWords, loc.ExcludeWords, err =
		cmd.QueryParser(receiveValue); err != nil { // 検索文字チェックERROR
		log.Errorf("%s [ %-50s ]", err, receiveValue)
		fmt.Fprint(w, htmlClause(receiveValue))
		fmt.Fprintf(w, ` %s
						</h4>
					<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.5.1/jquery.min.js"></script>
					<script type="text/javascript" src="js/datalist.js"></script>
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
			if err != nil {
				log.Error(err)
			}
			stats.Items = cmd.Ambiguous(n)
			stats.LastUpdateTime = DBLastUpdateTime()
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

		// Search result page
		fmt.Fprint(w, htmlClause(receiveValue))

		// Google検索の表示例
		// 約 8,010,000 件 （0.50 秒）
		fmt.Fprintf(w,
			`ヒット数: %d件中、最大%d件を表示<br>
			 %.3fmsec で約%s件を検索しました。<br>
			</h4>`,
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
