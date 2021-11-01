package locater

type (
	// CacheMap is normalized queries key and PathMap value pair
	CacheMap map[string]*Paths
)

// Traverse : 検索結果をcacheの中から探し、あれば検索結果と検索数を返し、
// なければLocater.Cmd()を走らせて検索結果と検索数を得る
func (cache *CacheMap) Traverse(l *Locater) (paths Paths, ok bool, err error) {
	s := l.Normalize() // Normlize for cache
	var v *Paths
	if v, ok = (*cache)[s]; !ok {
		// normalizedがcacheになければresultsをcacheに登録
		paths, err = l.Locate()
		(*cache)[s] = &paths
	} else {
		paths = *v
	}
	return
}
