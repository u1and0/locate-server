package locater

import (
	"log"

	"github.com/mattn/go-pipeline"
)

// LogParser : Log解析して検索語をチャネルに登録
func LogParser(f string) []string {
	pgstring := "(PUSH|GET) result (to|from) cache"
	pstring := "PUSH result to cache"
	gstring := "GET result from cache"
	grep := []string{"grep", "-oE", pgstring + ` \[.*\]`, f} // LOGから検索文字列抜き出し
	sed1 := []string{
		"sed", "-e", "s/^" + pstring + ` \[ //`, // PUSH...削除
		"-e", "s/^" + gstring + ` \[ //`, // GET...削除
		"-e", `s/\]$//`, // 最後の]削除
	}
	sort1 := []string{"sort"}
	uniq := []string{"uniq", "-c"}                // 重複カウント
	sort2 := []string{"sort", "-r"}               // 多い順に出力
	sed2 := []string{"sed", "-e", `s/^\s*//`}     // 行頭のスペースを削除
	cut := []string{"cut", "-d", ` `, "-f", "2-"} // カウント数削除
	/*
		logファイル内の検索ワードのみを抜き出し
		検索回数順に並び替えるshell script

		```shell
		grep -oE "(PUSH|GET).result.(to|from).cache.*\[.*\]" /var/lib/mlocate/locate.log |
			sed \
				-e "s/PUSH result to cache \[ //" \
				-e "s/GET result from cache \[ //" \
				-e "s/\]$//" |
			sort |
			uniq -c |
			sort -r |
			sed -e "s/^\s*\/\/" |
			cut -d " " -f 2-
		```
	*/

	out, err := pipeline.Output(grep, sed1, sort1, uniq, sort2, sed2, cut)
	if err != nil {
		log.Printf("[Fail] Log file parsing error out: %s, error: %s\n", out, err)
	}
	return SliceOutput(out)
}

// AutoCacheMaker : 自動キャッシュ生成
// &c のポインタで渡してキャッシュのメモリを直接書き換える(ので戻り値がない)
func (l *Locater) AutoCacheMaker(c *CacheMap, ch chan string) {
	var err error
	for {
		s, ok := <-ch
		if !ok {
			break
		}
		// channelから受け取った検索語を解析
		l.SearchWords, l.ExcludeWords, err = QueryParser(s)
		if err != nil {
			log.Printf("[Fail] Cache parsing error %s [ %-50s ] \n", err, s)
		}
		_, _, _, err = l.ResultsCache(c) // Cache生成
		if err != nil {
			log.Printf("[Fail] Making cache error %s [ %-50s ]\n", err, s)
		}
	}
}
