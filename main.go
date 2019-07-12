package main

import (
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
	results        []string
	dirs           []string
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

func showInit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html>
                       <head><title>Locate Server</title></head>
						 <body>
						   <form method="get" action="/searching">
							 <input type="text" name="query" value="%s">
							 <input type="submit" name="submit" value="検索">
						   </form>
						</body>
					</html>`, receiveValue)
}

// スペースを*に入れ替えて、前後に*を付与する
func patStar(s string) string {
	// s <= "hoge my name" のとき
	sn := strings.Fields(s)    // => [hoge my name]
	s = strings.Join(sn, ".*") // => hoge.*my.*name
	return s
}

// スライスのすべての要素の/を\に変換
func changeSepWin(sr []string) []string {
	for i, s := range sr {
		sr[i] = strings.ReplaceAll(s, "/", "\\") // Windows path
	}
	return sr
}

// root引数の文字列をスライスのすべての要素に追加
func addPrefix(sr []string) []string {
	for i, s := range sr {
		sr[i] = *root + s
	}
	return sr
}

func addResult(w http.ResponseWriter, r *http.Request) {
	// modify query
	receiveValue = r.FormValue("query")
	searchValue := patStar(receiveValue)
	log.Println("検索ワード:", receiveValue)

	// searching
	st := time.Now()
	out, err := exec.Command("locate", "-i", "--regex", searchValue).Output()
	if err != nil {
		log.Println(err)
	}
	en := time.Now()
	searchTime = (en.Sub(st)).Seconds()

	// mod results
	outstr := string(out)
	results = strings.Split(outstr, "\n")
	results = results[:len(results)-1] // Pop last element cause \\n

	// Dir path
	dirs = make([]string, len(results))
	for i, dir := range results {
		dirs[i] = filepath.Dir(dir)
	}

	// Change sep character / -> \
	if *pathSplitWin {
		results = changeSepWin(results)
		dirs = changeSepWin(dirs)
	}

	// Add network starge path to each of results
	if *root != "" {
		results = addPrefix(results)
		dirs = addPrefix(dirs)
	}

	// Max result 1000
	resultNum = len(results)
	if resultNum > 1000 {
		results = results[:1000]
	}
	log.Println("結果件数:", resultNum, "/", "検索時間:", searchTime)

	// update time
	fileStat, err := os.Stat("/var/lib/mlocate")
	layout := "2006-01-02 15:05"
	lastUpdateTime = fileStat.ModTime().Format(layout)
	if err != nil {
		log.Println(err)
	}

	// Search result page
	fmt.Fprintf(w, `<html>
                       <head><title>Locate Server</title></head>
						 <body>
						   <form method="get" action="/searching">
							 <input type="text" name="query" value="%s">
							 <input type="submit" name="submit" value="検索">
						   </form>`, receiveValue)
	receiveValue = "" // Reset form

	fmt.Fprintf(w, `<h4>
						 DB last update: %s<br>
						 検索結果          : %d件中、最大1000件を表示<br>
						 検索にかかった時間: %.3fsec
					</h4>`, lastUpdateTime, resultNum, searchTime)

	// 検索結果を行列表示
	fmt.Fprintln(w, `<table>
					  <tr>`)
	for i, rs := range results {
		fmt.Fprintf(w, `<tr>
		<td>
			<a href="file://%s">%s</a>
			<a href="file://%s" title="<< クリックでフォルダに移動"><<</a>
		</td>
	</tr>`, rs, rs, dirs[i])
	}

	fmt.Fprintln(w, `</table>
				  </body>
				  </html>`)
}
