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
	http.HandleFunc("/results", showResult)
	http.HandleFunc("/results/new", addResult)
	http.ListenAndServe(":8080", nil)
}

func showResult(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<html>")
	fmt.Fprintln(w, "<head><title>Locater</title></head>")

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h1>Locater</h1>")

	fmt.Fprintln(w, "<h2>検索</h2>")
	fmt.Fprintln(w, `<form method="post" action="/results/new">`)
	fmt.Fprintln(w, `<input type="text" name="query">`)
	fmt.Fprintln(w, `<input type="submit" name="submit">`)
	fmt.Fprintln(w, `</form>`)

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
	http.Redirect(w, r, "/results", 303)
}
