package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

var results []string

func main() {
	results = make([]string, 0)
	http.HandleFunc("/", showResult)
	http.HandleFunc("/searching", addResult)
	http.ListenAndServe(":8080", nil)
}

func showResult(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<html>")
	fmt.Fprintln(w, "<head><title>Locater</title></head>")

	fmt.Fprintln(w, "<body>")

	fmt.Fprintln(w, `<form method="post" action="/searching">`)
	fmt.Fprintln(w, `<input type="text" name="query">`)
	fmt.Fprintln(w, `<input type="submit" name="submit" value="検索">`)
	fmt.Fprintln(w, `</form>`)
	fmt.Fprintln(w, "<h4>検索結果: {.Name}件</h4>")

	fmt.Fprintln(w, "<table>")
	for _, result := range results {
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, "<td>%s</td>", result)
		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintln(w, "</table>")

	fmt.Fprintln(w, "</body>")
	fmt.Fprintln(w, "</html>")
}

func addResult(w http.ResponseWriter, r *http.Request) {
	receiveValue := r.FormValue("query")
	out, err := exec.Command("locate", receiveValue).Output()
	if err != nil {
		fmt.Println(err)
	}
	outstr := string(out)
	resultAll := strings.Split(outstr, "\n")
	results = resultAll[:1000]
	fmt.Println(results)
	http.Redirect(w, r, "/", 303)
}
