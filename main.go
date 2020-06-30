package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	cmd "locate-server/cmd"
)

const (
	// VERSION : version
	VERSION = "1.1.0"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// LOCATEDIR : locateのデータベースやログファイルを置く場所
	LOCATEDIR = "/var/lib/mlocate"
)

var (
	showVersion  bool
	receiveValue string
	err          error
	limit        int
	dbpath       string
	pathSplitWin bool
	root         string
	trim         string
	cache        cmd.CacheMap
	getpushLog   string
	locateS      []byte
	stats        cmd.Stats
)

func main() {
	flag.IntVar(&limit, "l", 1000, "Maximum limit for results")
	locatePath := os.Getenv("LOCATE_PATH")
	flag.StringVar(&dbpath, "d", locatePath, "path of locate database file (ex: /var/lib/mlocate/something.db)")
	flag.BoolVar(&pathSplitWin, "s", false, "OS path split windows backslash")
	flag.StringVar(&root, "r", "", "DB insert prefix for directory path")
	flag.StringVar(&trim, "t", "", "DB trim prefix for directory path")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()
	if showVersion {
		fmt.Println("version:", VERSION)
		return // versionを表示して終了
	}

	// Command check
	if _, err := exec.LookPath("locate"); err != nil {
		log.Fatal(err)
	}

	// Directory check
	if _, err := os.Stat(LOCATEDIR); os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("[ERROR] Cannot open logfile " + err.Error())
	}
	defer logfile.Close()
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	// Initialize cache
	// nil map assignment errorを発生させないために必要
	cache = cmd.CacheMap{}
	// cacheを廃棄するかの判断に必要
	// lstatが変わった=mlocate.dbの内容が更新されたのでcacheを新しくする
	locateS, err = cmd.LocateStats(dbpath)
	if err != nil {
		log.Println(err)
	}
	var n uint64
	n, err = cmd.LocateStatsSum(locateS)
	if err != nil {
		log.Println(err)
	}
	stats.Items = cmd.Ambiguous(n)
	if err != nil {
		log.Println(err)
	}

	// HTTP pages
	http.HandleFunc("/", showInit)
	http.HandleFunc("/searching", addResult)
	http.HandleFunc("/status", locateStatusPage)
	http.ListenAndServe(":8080", nil)
}

// html デフォルトの説明文
func htmlClause(s string) string {
	return fmt.Sprintf(`<html>
					<head><title>Locate Server %s</title></head>
					<body>
						<form method="get" action="/searching">
							<input type="text" name="query" value="%s" size="50">
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
						</small>`, s, s)
}

// Top page
func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause(""))
}

// `locate -S` page
func locateStatusPage(w http.ResponseWriter, r *http.Request) {
	// lとerrはstring型とerr型で異なるのでif-elseが冗長になる
	if l, err := cmd.LocateStats(dbpath); err == nil {
		fmt.Fprintf(w, `<html>
					<head><title>Locate DB Status</title></head>
					<body>
						<pre>%s</pre>
					</body>
					</html>`, l)
	} else {
		fmt.Fprintf(w, `<html>
						<head><title>Locate DB Status</title></head>
						<body>
							<pre>%s</pre>
						</body>
						</html>`, err)
	}
}

// locate検索し、結果をhtmlに書き込む
func addResult(w http.ResponseWriter, r *http.Request) {
	var (
		results []cmd.PathMap
	)
	// 検索コンフィグ構造体
	loc := cmd.Locater{
		Limit:        limit,        // 検索件数上限
		Dbpath:       dbpath,       // /var/lib/mlocate以外のディレクトリパス
		PathSplitWin: pathSplitWin, // path separatorを\にする
		Root:         root,         // Path prefix insert
		Trim:         trim,         // Path prefix trim
	}
	// Modify query
	receiveValue = r.FormValue("query")

	if loc.SearchWords, loc.ExcludeWords, err =
		cmd.QueryParser(receiveValue); err != nil { // 検索文字チェックERROR
		log.Printf("%s [ %-50s ] \n", err, receiveValue)
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
				log.Println(err)
			}
			locateS = l // 保持するDB情報の更新
			var n uint64
			n, err = cmd.LocateStatsSum(l) // 検索ファイル数の更新
			stats.Items = cmd.Ambiguous(n)
			if err != nil {
				log.Println(err)
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
			log.Printf("%s [ %-50s ]\n", err, receiveValue)
		}
		log.Printf("%8dfiles %3.3fmsec %s [ %-50s ]\n",
			stats.ResultNum, stats.SearchTime, getpushLog, receiveValue)
		/* normalizedWordではなく、あえてreceiveValueを
		表示して生の検索文字列を記録したい*/

		// Update time
		filestat, err := os.Stat(LOCATEDIR)
		if err != nil {
			log.Println(err)
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
