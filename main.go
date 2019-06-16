package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

var (
	results   []string
	resultNum int
)

func main() {
	results = make([]string, 0)
	resultNum = 0
	http.HandleFunc("/", showResult)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}

func showResult(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html>
                       <head><title>Locate Server</title></head>
						 <body>
						   <form method="post" action="/searching">
							 <input type="text" name="query">
							 <input type="submit" name="submit" value="検索">
						   </form>`)

	fmt.Fprintf(w, `<form method="post" action="/searching">
						 <h4>検索結果: %d件中、最大1000件を表示</h4>
					 </form>`, resultNum)

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

func addResult(w http.ResponseWriter, r *http.Request) {
	receiveValue := r.FormValue("query")
	out, err := exec.Command("locate", receiveValue).Output()
	if err != nil {
		fmt.Println(err)
	}
	outstr := string(out)
	resultAll := strings.Split(outstr, "\n")
	resultNum = len(resultAll)
	results = resultAll[:1000]
	fmt.Println("検索ワード:", receiveValue, "結果件数:", resultNum)
	http.Redirect(w, r, "/", 303)
}
