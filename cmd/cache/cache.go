package cache

import (
	cmd "github.com/u1and0/locate-server/cmd/locater"
)

type (
	// Map is normalized queries key and PathMap value pair
	Map map[Key]*cmd.Paths

	// Key : Key for Map
	Key struct {
		Word  string
		Limit int
	}
)

// New : cache constructor
func New() *Map {
	return &Map{}
}

// Traverse : 検索結果をcacheの中から探し、あれば検索結果と検索数を返し、
// なければLocater.Cmd()を走らせて検索結果と検索数を得る
func (cache *Map) Traverse(l *cmd.Locater) (paths cmd.Paths, ok bool, err error) {
	s := Key{cmd.Normalize(l.SearchWords, l.ExcludeWords), l.Query.Limit}
	var v *cmd.Paths
	if v, ok = (*cache)[s]; !ok {
		// normalizedがcacheになければresultsをcacheに登録
		paths, err = l.Locate()
		(*cache)[s] = &paths
	} else {
		paths = *v
	}
	return
}
