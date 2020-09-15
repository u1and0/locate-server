package locater

import (
	"bufio"
	"io"
	"os"
	"strings"
)

const (
	// LOGFILE : 検索条件 / 検索結果 / 検索時間を記録するファイル
	LOGFILE = "/var/lib/mlocate/locate.log"
)

// LogWord extract search word from LOGFILE
func LogWord() (ss []string, err error) {
	fp, err := os.Open(LOGFILE)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)

	for {
		var line []byte
		line, _, err = reader.ReadLine()
		if err == io.EOF { // if EOF then finish func
			err = nil
			return
		}
		if err != nil {
			return
		}
		lines := string(line)
		// start := bytes.IndexByte(line, []byte(`[ `))
		// end := bytes.IndexByte(line, []byte(` ]`))
		start := strings.Index(lines, "[ ")
		end := strings.Index(lines, " ]")
		if start > -1 {
			searchWord := lines[start+1 : end-1]
			ss = append(ss, searchWord)
		}
	}
}
