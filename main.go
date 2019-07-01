package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
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
	root           = flag.String("r", "", "DB root directory")
	pathSplitWin   = flag.Bool("s", false, "OS path split windows backslash")
)

func main() {
	resultNum = 0

	flag.Parse()

	http.HandleFunc("/", showResult)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}

func showResult(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html>
                       <head><title>Locate Server</title></head>
						 <body>
						   <form method="post" action="/searching">
							 <input type="text" name="query" value="%s">
							 <input type="submit" name="submit" value="検索">
						   </form>`, receiveValue)

	fmt.Fprintf(w, `<form method="post" action="/searching">
						 <h4>
						 DB last update: %s<br>
						 検索結果          : %d件中、最大1000件を表示<br>
						 検索にかかった時間: %.3fsec
						 </h4>
					 </form>`, lastUpdateTime, resultNum, searchTime)

	// 検索結果を行列表示
	fmt.Fprintln(w, `<table>
					  <tr>`)
	for i, r := range results {
		fmt.Fprintf(w, `<tr>
		<td>
			<a href="file://%s">%s</a>
			<a href="file://%s" title="<< クリックでフォルダに移動"><<</a>
		</td>
						</tr>`, r, r, dirs[i])
	}

	fmt.Fprintln(w, `</table>
				  </body>
				  </html>`)
}

// スペースを*に入れ替えて、前後に*を付与する
func patStar(s string) string {
	// s <= "hoge my name" のとき
	sn := strings.Fields(s)   // => [hoge my name]
	s = strings.Join(sn, "*") // => hoge*my*name
	s = "*" + s + "*"         // => *hoge*my*name*
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

	// Query encoding
	params := url.Values{}
	params.Add("q", receiveValue)
	que := params.Encode()
	fmt.Println(que)
	fmt.Println("検索ワード:", receiveValue)

	searchValue := patStar(receiveValue)

	// searching
	st := time.Now()
	out, err := exec.Command("locate", "-i", searchValue).Output()
	if err != nil {
		fmt.Println(err)
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
	fmt.Println("結果件数:", resultNum, "/", "検索時間:", searchTime)

	// update time
	fileStat, err := os.Stat("/var/lib/mlocate")
	layout := "2006-01-02 15:05"
	lastUpdateTime = fileStat.ModTime().Format(layout)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/", 303)
}
