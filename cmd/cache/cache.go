package cache

import (
	cmd "github.com/u1and0/locate-server/cmd/locater"
)

type (
	// Map is normalized queries key and PathMap value pair
	Map map[Key]*cmd.Paths
	// Key : cache Map key
	Key struct {
		Word  string // Normalized query
		Limit int    // Number of results
	}
)

// New : cache constructor
func New() *Map {
	return &Map{}
}

// Traverse : 検索結果をcacheの中から探し、あれば検索結果と検索数を返し、
// なければLocater.Cmd()を走らせて検索結果と検索数を得る
func (cache *Map) Traverse(l *cmd.Locater) (paths cmd.Paths, ok bool, err error) {
	w := cmd.Normalize(l.SearchWords, l.ExcludeWords)
	k := Key{w, l.Query.Limit}
	if v, ok := (*cache)[k]; !ok {
		// normalizedがcacheになければresultsをcacheに登録
		paths, err = l.Locate()
		(*cache)[k] = &paths
	} else {
		paths = *v
	}
	return
}
