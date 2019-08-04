package locater

import "log"

type (
	// CacheMap is normalized queries key and PathMap value pair
	CacheMap map[string]*CacheStruct
)

// CacheStruct : 検索クエリをキーとした検索結果と検索数を保管したキャッシュ
type CacheStruct struct {
	Paths []PathMap
	Num   int
}

// ResultsCache : 検索結果をcacheの中から探し、あれば検索結果と検索数を返し、
// なければLocater.Cmd()を走らせて検索結果と検索数を得る
func (l *Locater) ResultsCache(cache CacheMap) ([]PathMap, int, CacheMap, error) {
	var (
		results   []PathMap
		resultNum int
		err       error
	)
	normalized := l.Normalize() // Normlize for cache
	if ce, ok := cache[normalized]; !ok {
		// normalizedがcacheになければresultsとresultNumをcacheに登録
		results, resultNum, err = l.Cmd()
		cache[normalized] = &CacheStruct{
			Paths: results,
			Num:   resultNum,
		}
		log.Printf("[ %-50s ] result PUSH to cache\n", normalized)
	} else {
		// normalizedがcacheにあればcacheからresultsとresultNumを取り出す
		results = ce.Paths
		resultNum = ce.Num
		log.Printf("[ %-50s ] result GET from cache\n", normalized)
	}
	return results, resultNum, cache, err
}