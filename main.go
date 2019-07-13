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
	"strings"
	"time"
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
)

func main() {
	flag.Parse()

	// Log setting
	logfile, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannot open logfile" + err.Error())
	}
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))

	// HTTP pages
	http.HandleFunc("/", showInit)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}

func htmlClause(s string) string {
	return fmt.Sprintf(`<html>
					<head><title>Locate Server</title></head>
					<body>
						<form method="get" action="/searching">
							<input type="text" name="query" value="%s">
							<input type="submit" name="submit" value="検索">
						</form>
						<p>
							 * 対象文字列は2文字以上の文字列を指定してください。<br>
							 * スペース区切りで複数入力できます。(AND検索)<br>
							 * 半角カッコでくくって|で区切ると|で区切られる前後で検索します。(OR検索)<br>
							 例: "電(気|機)工業" => "電気工業"と"電機工業"を検索します。
						</p>`, s)
}

func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlClause(receiveValue))
}

// スペースを*に入れ替えて、前後に*を付与する
func patStar(s string) (string, error) {
	var (
		sn  []string
		err error
	)
	// s <= "hoge my name" のとき
	if len([]rune(s)) < 2 {
		err = errors.New("検索文字列が足りません")
	} else {
		sn = strings.Fields(s)     // => [hoge my name]
		s = strings.Join(sn, ".*") // => hoge.*my.*name
	}
	return s, err
}

func addResult(w http.ResponseWriter, r *http.Request) {
	// Modify query
	receiveValue = r.FormValue("query")
	log.Println("検索ワード:", receiveValue)
	if searchValue, err := patStar(receiveValue); err != nil { // 検索文字列が1文字以下のとき
		log.Println(err)
		fmt.Fprint(w, htmlClause(receiveValue))
		fmt.Fprintln(w, `<h4>
							検索文字列が足りません
						</h4>
					</body>
					</html>`)
	} else {
		// Searching
		st := time.Now()
		out, err := exec.Command("locate", "-i", "--regex", searchValue).Output()
		if err != nil {
			log.Println(err)
		}
		en := time.Now()
		searchTime = (en.Sub(st)).Seconds()

		// Mod results
		results = make(map[string]string, 1000000)
		for _, f := range strings.Split(string(out), "\n") {
			results[f] = filepath.Dir(f)
		}
		delete(results, "") // Pop last element cause \\n

		// Change sep character / -> \
		if *pathSplitWin { // Windows path
			for k, v := range results {
				delete(results, k)
				results[strings.ReplaceAll(k, "/", "\\")] = strings.ReplaceAll(v, "/", "\\")
			}
		}

		// Add network starge path to each of results
		if *root != "" {
			for k, v := range results {
				delete(results, k)
				results[*root+k] = *root + v
			}
		}

		resultNum = len(results)
		log.Println("結果件数:", resultNum, "/", "検索時間:", searchTime)

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
							 DB last update: %s<br>
							 検索結果          : %d件中、最大1000件を表示<br>
							 検索にかかった時間: %.3fsec
						</h4>`, lastUpdateTime, resultNum, searchTime)

		// 検索結果を行列表示
		fmt.Fprintln(w, `<table>
						  <tr>`)
		i := 1
		for f, d := range results {
			if i++; i > 1000 { // Max results 1000
				break
			}
			fmt.Fprintf(w, `<tr>
				<td>
					<a href="file://%s">%s</a>
					<a href="file://%s" title="<< クリックでフォルダに移動"><<</a>
				</td>
			</tr>`, f, f, d)
		}

		fmt.Fprintln(w, `</table>
					  </body>
					  </html>`)
	}
}
