package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	cmd "locate-server/cmd"

	cache "github.com/patrickmn/go-cache"
)

const (
	// VERSION : version
	VERSION = "1.0.3"
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// CAP : 表示する検索結果上限数
	CAP = 1000
	// LOCATEPATH : locateのデータベースやログファイルを置く場所
	LOCATEPATH = "/var/lib/mlocate"
	//PROCESSES : locateをキャッシュ化するときの並列プロセス数
	PROCESSES = 1
)

var (
	showVersion  bool
	receiveValue string
	err          error
	root         = flag.String("r", "", "DB root directory")
	pathSplitWin = flag.Bool("s", false, "OS path split windows backslash")
	dbpath       = flag.String("d", "", "path of locate database file (ex: /var/lib/mlocate/something.db)")
	cacheMap     = cache.New(cache.NoExpiration, cache.DefaultExpiration)
	results      []cmd.PathMap
	resultNum    int
	getpushLog   string
	lstatinit    []byte
)

func main() {
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
	if _, err := os.Stat("/var/lib/mlocate"); os.IsNotExist(err) {
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

	// cache = *cmd.NewCacheMap()
	/* cacheを廃棄するかの判断に必要
	lstatが変わった=mlocate.dbの内容が更新されたのでcacheを新しくする */

	/* N秒おきにlocatestat()でdbの状態チェックして、
	変更あればautocache()でcacheの再構成 */
	go func() {
		for {
			/* locatestat()の結果が前と異なっていたら
			lstatinit更新
			cacheを再構成 */
			l, err := locatestat()
			if err != nil {
				log.Println(err)
			}
			if string(l) != string(lstatinit) {
				lstatinit = l
				cacheMap = cache.New(cache.NoExpiration, cache.DefaultExpiration)
				// cacheを廃棄するかの判断に必要
				// lstatが変わった=mlocate.dbの内容が更新されたのでcacheを新しくする
				lstatinit, err = locatestat()
				if err != nil {
					log.Println(err)
				}
				autocache()
			}
			time.Sleep(10 * time.Second)
		}
	}()

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

// Result of `locate -S`
func locatestat() ([]byte, error) {
	opt := []string{"-S"}
	if *dbpath != "" {
		opt = append(opt, "-d", *dbpath)
	}
	return exec.Command("locate", opt...).Output()
}

// `locate -S` page
func locateStatusPage(w http.ResponseWriter, r *http.Request) {
	// lとerrはstring型とerr型で異なるのでif-elseが冗長になる
	if l, err := locatestat(); err == nil {
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
	// Modify query
	receiveValue = r.FormValue("query")
	loc := cmd.Locater{
		Dbpath:       *dbpath,       // /var/lib/mlocate以外のディレクトリパス
		Cap:          CAP,           // 検索件数上限
		PathSplitWin: *pathSplitWin, // path separatorを\にする
		Root:         *root,         // Path prefix
	}

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
		/* locatestat()の結果が前と異なっていたら
		lstatinit更新
		cacheを初期化 */
		if l, err := locatestat(); string(l) != string(lstatinit) {
			if err != nil {
				log.Println(err)
			} else {
				lstatinit = l
				cacheMap.Flush()
				cacheMap = cache.New(cache.NoExpiration, cache.DefaultExpiration)
			}
		}

		// Searching
		startTime := time.Now()
		results, resultNum, getpushLog, err = loc.ResultsCache(cacheMap)
		/* cache は&cacheによりdeep copyされてResultsCache()内で
		直接書き換えられるので、returnされない*/
		elapsed := time.Since(startTime)
		// nano sec 変換
		searchTime := float64(elapsed.Nanoseconds()) / float64(time.Millisecond)

		if err != nil {
			log.Printf("%s [ %-50s ]\n", err, receiveValue)
		}
		log.Printf("%8dfiles %3.3fmsec %s [ %-50s ]\n",
			resultNum, searchTime, getpushLog, receiveValue)
		/* normalizedWordではなく、あえてreceiveValueを
		表示して生の検索文字列を記録したい*/

		// Update time
		filestat, err := os.Stat(LOCATEPATH)
		if err != nil {
			log.Println(err)
		}
		layout := "2006-01-02 15:05"
		lastUpdateTime := filestat.ModTime().Format(layout)

		// Search result page
		fmt.Fprint(w, htmlClause(receiveValue))

		fmt.Fprintf(w, `<h4>
							 <a href=/status>DB</a> last update: %s<br>
							 検索結果          : %d件中、最大1000件を表示<br>
							 検索にかかった時間: %.3fmsec
						</h4>`, lastUpdateTime, resultNum, searchTime)

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

// autocache : Cacheを自動生成する
// logを解析して検索語をchannelに送信し、Cacheを自動生成する
//
// ! 正しいワードにHighlightが格納されていない
//
func autocache() {
	ch := make(chan *cmd.Locater)
	defer close(ch)
	loc := cmd.Locater{
		Dbpath:       *dbpath,
		Cap:          CAP,
		PathSplitWin: *pathSplitWin,
		Root:         *root,
	}

	// PROCESSES(デフォルト4)並列処理でキャッシュを作成
	for i := 0; i < PROCESSES; i++ {
		go func() {
			for {
				loch, ok := <-ch
				if !ok {
					break
				}
				fmt.Printf("[INFO] Try to make chach [ %s ]\n", loch.Normalize())
				// なぜか2回表示される

				/* 元のmainで解析
				// channelから受け取った検索語を解析
				l.SearchWords, l.ExcludeWords, err = QueryParser(s)
				if err != nil {
					log.Printf("[Fail] Cache parsing error %s [ %-50s ] \n", err, s)
				}
				*/
				_, _, _, err = loch.ResultsCache(cacheMap) // Cache生成
				if err != nil {
					log.Printf("[Fail] Making cache error %s [ %-50s ]\n",
						err, loch.Normalize())
				}
			}

		}()
	}

	for _, q := range cmd.LogParser(LOGFILE) {
		loc.SearchWords, loc.ExcludeWords, err = cmd.QueryParser(q)
		if err != nil {
			log.Printf("[Fail] Cache parsing error %s [ %-50s ] \n", err, q)
		}
		ch <- &loc
	}

	log.Printf("[INFO] Cached words [ %s ]\n",
		strings.Join(func() (s []string) {
			for k := range cacheMap.Items() { // cache化に成功した語を表示
				s = append(s, k)
			}
			return
		}(), ", "))
}
