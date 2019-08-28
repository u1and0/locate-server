package locater

import cache "github.com/patrickmn/go-cache"

// CacheStruct : 検索クエリをキーとした検索結果と検索数を保管したキャッシュ
type CacheStruct struct {
	Paths []PathMap
	Num   int
}

// ResultsCache : 検索結果をcacheの中から探し、あれば検索結果と検索数を返し、
// なければLocater.Cmd()を走らせて検索結果と検索数を得る
func (l *Locater) ResultsCache(c *cache.Cache) ([]PathMap, int, string, error) {
	normalized := l.Normalize() // Normlize for cache
	ce, found := c.Get(normalized)
	if !found {
		// normalizedがcacheになければresultsとresultNumをcacheに登録
		results, resultNum, err := l.Cmd()
		c.Set(normalized,
			&CacheStruct{
				Paths: results,
				Num:   resultNum,
			},
			cache.NoExpiration)
		return results, resultNum, "PUSH result to cache", err
	}
	// normalizedがcacheにあればcacheからresultsとresultNumを取り出す
	return ce.(*CacheStruct).Paths, ce.(*CacheStruct).Num, "GET result from cache", nil
}
