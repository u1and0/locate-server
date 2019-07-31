package locater

import (
	"path/filepath"
	"strings"

	pipeline "github.com/mattn/go-pipeline"
)

type (
	// PathMap is pairs of fullpath:dirpath
	PathMap map[string]string
)

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

// Cmd : locate検索し、結果をfullpath:dirpathのマップ(最大capacity件)にして返す
// 更に検索結果数、あれば検索時のエラーを返す
func (l *Locater) Cmd(capacity int) (PathMap, int, error) {
	out, err := pipeline.Output(l.CmdGen()...)
	outslice := strings.Split(string(out), "\n")
	outslice = outslice[:len(outslice)-1] // Pop last element cause \\n

	// Map parent directory name
	results := make(PathMap, capacity)
	var i int
	for _, f := range outslice {
		if i++; i > capacity {
			break
		}
		results[f] = filepath.Dir(f)
	}

	return results, len(outslice), err // Max 1000 result & number of all result
}
