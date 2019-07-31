package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	cmd "locate-server/cmd"
)

const (
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
	// CAP : 表示する検索結果上限数
	CAP = 1000
	// LOCATEPATH : locateのデータベースやログファイルを置く場所
	LOCATEPATH = "/var/lib/mlocate"
)

var (
	results        cmd.PathMap
	cache          CacheMap
	lastUpdateTime string
	searchTime     float64
	receiveValue   string
	err            error
	root           = flag.String("r", "", "DB root directory")
	pathSplitWin   = flag.Bool("s", false, "OS path split windows backslash")
	dbpath         = flag.String("d", "", "path of locate database file (ex: /var/lib/mlocate/something.db)")
)

type (
	// CacheMap is normalized queries key and PathMap value pair
	CacheMap map[string]cmd.PathMap
)

func main() {
	flag.Parse()

	// Log setting
	logfile, err := os.OpenFile(LOGFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("[warning] cannot open logfile" + err.Error())
	}
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	// HTTP pages
	http.HandleFunc("/", showInit)
	http.HandleFunc("/searching", addResult)
	http.HandleFunc("/status", locateStatus)
	http.ListenAndServe(":8080", nil)
}

// html デフォルトの説明文
func htmlClause(s string) string {
	return fmt.Sprintf(`<html>
					<head><title>Locate Server</title></head>
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
						</small>`, s)
}

// Top page
func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause(receiveValue))
}

// スペースを*に入れ替えて、前後に*を付与する
func queryParser(s string) (sn, en []string, err error) {
	// s <- "hoge my -your name"
	for _, n := range strings.Fields(s) { // -> [hoge my -your name]
		if strings.HasPrefix(n, "-") {
			en = append(en, strings.TrimPrefix(n, "-")) // ->[your]
		} else {
			sn = append(sn, n) // ->[hoge my name]
		}
	}
	if len([]rune(strings.Join(sn, ""))) < 2 {
		err = errors.New("検索文字数が足りません")
	}
	return
}

// Result of `locate -S`
func locateStatus(w http.ResponseWriter, r *http.Request) {
	opt := []string{"-S"}
	if *dbpath != "" {
		opt = append(opt, "-d", *dbpath)
	}

	locates, err := exec.Command("locate", opt...).Output()
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, `<html>
					<head><title>Locate DB Status</title></head>
					<body>
						<pre>%s</pre>
					</body>
					</html>`, locates)
}

// sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words []string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`)
		s = re.ReplaceAllString(s, "<span style=\"background-color:#FFCC00;\">$1</span>")
	}
	return s
}

// locate検索し、結果をhtmlに書き込む
func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	receiveValue = r.FormValue("query")
	loc := new(cmd.Locater)
	loc.Dbpath = *dbpath
	if loc.SearchWords, loc.ExcludeWords, err = queryParser(receiveValue); err != nil { // 検索文字列が1文字以下のとき
		log.Println(err)
		fmt.Fprint(w, htmlClause(receiveValue))
		fmt.Fprintln(w, `<h4>
							検索文字数が足りません
						</h4>
					</body>
					</html>`)
	} else {
		// Normlized word for cache
		normalizeWord := loc.Normalize()

		// Searching
		startTime := time.Now()
		results, resultNum, err := loc.Cmd(CAP)
		searchTime = (time.Since(startTime)).Seconds()

		// Change sep character / -> \
		if *pathSplitWin { // Windows path
			r := make(map[string]string, 10000)
			for k, v := range results {
				r[strings.ReplaceAll(k, "/", "\\")] = strings.ReplaceAll(v, "/", "\\")
			}
			results = r
		}

		// Add network starge path to each of results
		if *root != "" {
			r := make(map[string]string, 10000)
			for k, v := range results {
				r[*root+k] = *root + v
			}
			results = r
		}

		log.Printf("検索ワード: %-40s 結果件数:%8d 検索時間: %3.3f\n",
			normalizeWord, resultNum, searchTime)

		// Update time
		filestat, err := os.Stat(LOCATEPATH)
		if err != nil {
			log.Println(err)
		}
		layout := "2006-01-02 15:05"
		lastUpdateTime = filestat.ModTime().Format(layout)

		// Search result page
		fmt.Fprint(w, htmlClause(receiveValue))
		receiveValue = "" // Reset form
		// これがないと次回アクセス時の最初のページ8080のフォームが最後検索した文字列になる

		fmt.Fprintf(w, `<h4>
							 <a href=/status>DB</a> last update: %s<br>
							 検索結果          : %d件中、最大1000件を表示<br>
							 検索にかかった時間: %.3fsec
						</h4>`, lastUpdateTime, resultNum, searchTime)

		// 検索結果を行列表示
		fmt.Fprintln(w, `<table>
						  <tr>`)
		for f, d := range results {
			fmt.Fprintf(w, `<tr>
				<td>
					<a href="file://%s">%s</a>
					<a href="file://%s" title="<< クリックでフォルダに移動"><<</a>
				</td>
			</tr>`, f, highlightString(f, loc.SearchWords), d)
		}

		fmt.Fprintln(w, `</table>
					  </body>
					  </html>`)
	}
}
