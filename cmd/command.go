package locater

import (
	"path/filepath"
	"regexp"
	"strings"

	pipeline "github.com/mattn/go-pipeline"
)

// PathMap is pairs of fullpath:dirpath
type PathMap struct {
	File      string
	Dir       string
	Highlight string
}

// sの文字列中にあるwordsの背景を黄色にハイライトしたhtmlを返す
func highlightString(s string, words []string) string {
	for _, w := range words {
		re := regexp.MustCompile(`((?i)` + w + `)`)
		s = re.ReplaceAllString(s, "<span style=\"background-color:#FFCC00;\">$1</span>")
	}
	return s
}

// CmdGen : locate 検索語 | grep -v 除外語 | grep -v 除外語...を発行する
func (l *Locater) CmdGen() [][]string {
	locate := []string{"locate", "-i"} // -i: Ignore case distinctions when matching patterns.
	if l.Dbpath != "" {
		locate = append(locate, "-d", l.Dbpath) // -d: Replace the default database with DBPATH.
	}

	// Include PATTERNs
	locate = append(locate, "--regex", strings.Join(l.SearchWords, ".*"))
	// -> hoge.*my.*name

	// Exclude PATTERNs
	exec := [][]string{locate}
	for _, ex := range l.ExcludeWords {
		exec = append(exec, []string{"grep", "-ivE", ex})
	}
	return exec
}

// Cmd : locate検索し、結果をPathMapのスライス(最大l.Cap件(capacity = default 1000))にして返す
// 更に検索結果数、あれば検索時のエラーを返す
func (l *Locater) Cmd() ([]PathMap, int, error) {
	out, err := pipeline.Output(l.CmdGen()...)
	outslice := strings.Split(string(out), "\n")
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n

	// Map parent directory name
	results := make([]PathMap, 0, l.Cap)
	for i, f := range outslice {
		// l.Cap  件までresultsとして返す
		if i >= l.Cap {
			break
		}
		results = append(results, PathMap{
			f,
			filepath.Dir(f),
			highlightString(f, l.SearchWords),
		})
	}

	// Windows path
	if l.PathSplitWin {
		for i, p := range results {
			results[i] = p.ChangeSep("\\", l.SearchWords)
		}
	}
	// Add network starge path to each of results
	if l.Root != "" {
		for i, p := range results {
			results[i] = p.AddPrefix(l.Root)
		}
	}

	return results, len(outslice), err // Max 1000 result & number of all result
}

// ChangeSep : Change file path separator to arbitrary s
func (p *PathMap) ChangeSep(s string, searchwords []string) PathMap {
	f := strings.ReplaceAll(p.File, "/", "\\")
	d := strings.ReplaceAll(p.Dir, "/", "\\")
	h := highlightString(f, searchwords)
	return PathMap{f, d, h}
}

// AddPrefix : Change file path separator to arbitrary s
func (p *PathMap) AddPrefix(r string) PathMap {
	return PathMap{r + p.File, r + p.Dir, r + p.Highlight}
}
