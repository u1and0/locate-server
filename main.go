package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const dbroot = "/var/lib/mlocate/"

var (
	results        []string
	resultNum      int
	lastUpdateTime string
	dbpath         string
)

func main() {
	results = make([]string, 0)
	resultNum = 0
	dbpath = ""
	http.HandleFunc("/", showResult)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}

func showResult(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html>
                       <head><title>Locate Server</title></head>
						 <body>
						   <form method="post" action="/searching">
						   <table>
							   <tr>
								 <td><select name="dbpath-name">`)

	// dbroot以下のファイルを表示し、検索データベースを選択する
	files, err := filepath.Glob(dbroot + "*")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Fprintf(w, `<option value="%s" selected>%s</option>`, f, f)
	}
	fmt.Fprintln(w, `   		 </select></td>
							   </tr>
						   </table>
							 <input type="text" name="query">
							 <input type="submit" name="submit" value="検索">
						   </form>`)

	fmt.Fprintf(w, `<form method="post" action="/searching">
						 <h4>DB last update: %s<br>
						 検索結果: %d件中、最大1000件を表示</h4>
					 </form>`, lastUpdateTime, resultNum)

	// 検索結果を行列表示
	fmt.Fprintln(w, `<table>
					  <tr>`)
	for _, result := range results {
		fmt.Fprintf(w, `<tr>
						  <td>%s</td>
						</tr>`, result)
	}

	fmt.Fprintln(w, `</table>
				  </body>
				  </html>`)
}

// スペースを*に入れ替えて、前後に*を付与する
func patStar(s string) string {
	// s <= "hoge my name" のとき
	sn := strings.Split(s, " ") // => [hoge my name]
	s = strings.Join(sn, "*")   // => hoge*my*name
	s = "*" + s + "*"           // => *hoge*my*name*
	return s
}

func addResult(w http.ResponseWriter, r *http.Request) {
	receiveValue := r.FormValue("query")
	receiveValue = patStar(receiveValue)
	dbpath = r.FormValue("dbpath-name")
	out, err := exec.Command("locate", "-id", dbpath, receiveValue).Output()
	fileStat, err := os.Stat(dbpath)
	layout := "2006-01-02 15:05"
	lastUpdateTime = fileStat.ModTime().Format(layout)
	if err != nil {
		fmt.Println(err)
	}
	outstr := string(out)
	results = strings.Split(outstr, "\n")
	resultNum = len(results) - 1 // \nのため、resultsの最後の要素は"空"になる
	if resultNum > 1000 {
		results = results[:1000]
	}

	// Log
	fmt.Printf("[%v] ", time.Now().Format(layout))
	fmt.Println("DB:", dbpath,
		"検索ワード:", receiveValue,
		"結果件数:", resultNum)

	http.Redirect(w, r, "/", 303)
}
