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
	nwrd := l.Normalize() // Normlize for cache
	if cacheElem, ok := cache[nwrd]; !ok {
		// nwrdがcacheになければresultsとresultNumをcacheに登録
		results, resultNum, err = l.Cmd()
		cache[nwrd] = &CacheStruct{
			Paths: results,
			Num:   resultNum,
		}
		log.Println("Result push to cache")
	} else {
		// cnwrdがacheにあればcacheからresults と　resultNumを取り出す
		results = cacheElem.Paths
		resultNum = cacheElem.Num
		log.Println("Result get from cache")
	}
	return results, resultNum, cache, err
}
