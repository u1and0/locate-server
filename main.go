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
	"path/filepath"
	"regexp"
	"strings"
	"time"

	pipeline "github.com/mattn/go-pipeline"
)

var (
	results        map[string]string
	resultNum      int
	lastUpdateTime string
	searchTime     float64
	receiveValue   string
	logfile        = "/var/lib/mlocate/locate.log"
	root           = flag.String("r", "", "DB root directory")
	pathSplitWin   = flag.Bool("s", false, "OS path split windows backslash")
	dbpath         = flag.String("d", "", "path of locate database file (ex: /var/lib/mlocate/something.db)")
)

func main() {
	flag.Parse()

	// Log setting
	logfile, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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
func patStar(s string) (sn, en []string, err error) {
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
	if searchWords, excludeWords, err := patStar(receiveValue); err != nil { // 検索文字列が1文字以下のとき
		log.Println(err)
		fmt.Fprint(w, htmlClause(receiveValue))
		fmt.Fprintln(w, `<h4>
							検索文字数が足りません
						</h4>
					</body>
					</html>`)
	} else {
		// Normlized word for cache
		normalizeWord := strings.Join(append(searchWords, excludeWords...), " ")

		// Search options
		cmd := []string{"locate", "-i"} // -i: Ignore case distinctions when matching patterns.
		if *dbpath != "" {
			cmd = append(cmd, "-d", *dbpath) // -d: Replace the default database with DBPATH.
		}
		// Interpret all PATTERNs as extended regexps.
		cmd = append(cmd, "--regex", strings.Join(searchWords, ".*"))
		// -> hoge.*my.*name

		// Exclude PATTERNs
		exes := [][]string{cmd} // locate cmd & piped cmd
		for _, ex := range excludeWords {
			exes = append(exes, []string{"grep", "-ivE", ex})
		}

		// Searching
		st := time.Now()
		out, err := pipeline.Output(exes...)
		if err != nil {
			log.Println(err)
		}
		en := time.Now()
		searchTime = (en.Sub(st)).Seconds()

		// Map parent directory name
		results = make(map[string]string, 10000)
		for _, f := range strings.Split(string(out), "\n") {
			results[f] = filepath.Dir(f)
		}
		delete(results, "") // Pop last element cause \\n

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

		resultNum = len(results)
		log.Println("検索ワード:", normalizeWord, "/", "結果件数:", resultNum, "/", "検索時間:", searchTime)

		// Update time
		fileStat, err := os.Stat("/var/lib/mlocate")
		layout := "2006-01-02 15:05"
		lastUpdateTime = fileStat.ModTime().Format(layout)
		if err != nil {
			log.Println(err)
		}

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
		i := 0
		for f, d := range results {
			if i++; i > 1000 { // Max results 1000
				break
			}
			fmt.Fprintf(w, `<tr>
				<td>
					<a href="file://%s">%s</a>
					<a href="file://%s" title="<< クリックでフォルダに移動"><<</a>
				</td>
			</tr>`, f, highlightString(f, searchWords), d)
		}

		fmt.Fprintln(w, `</table>
					  </body>
					  </html>`)
	}
}
