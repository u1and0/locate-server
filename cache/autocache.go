package locater

import (
	"io/ioutil"
	"log"
	"regexp"
)

// AutoCache : locate.logを解析してバックグラウンドでcacheを生成する
func AutoCache(file string) []string {
	// read file
	// f, err := os.OpenFile(file, os.O_RDONLY, 0666)
	// if err != nil {
	// 	log.Println("[warning] cannot open logfile" + err.Error())
	// }
	// defer f.Close()
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("[warning] cannot read logfile" + err.Error())
	}

	// regex
	re := regexp.MustCompile(`\[.*\]`)
	var line []string
	line = append(line, re.FindAllString(string(b), -1)...)
	return line

	// strigs
	// strings.HasPrefix(string(b).)
}
