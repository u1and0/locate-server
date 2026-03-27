package locater

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DBLastUpdateTime returns date time string for directory update time
func DBLastUpdateTime(db string) string {
	filestat, err := os.Stat(db)
	if err != nil {
		slog.Error("DBLastUpdateTime", "err", err)
	}
	layout := "2006-01-02 15:05"
	return filestat.ModTime().Format(layout)
}

// DBSize returns total byte size of locate database files.
// Used only for cache invalidation detection.
func DBSize(s string) (int64, error) {
	var sum int64
	dbs, err := filepath.Glob(s + "/*.db")
	if err != nil {
		return sum, err
	}
	for _, d := range dbs {
		file, err := os.Open(d)
		defer file.Close()
		i, err := file.Stat()
		s := i.Size()
		if err != nil {
			return s, err
		}
		sum += s
	}
	return sum, err
}

// Normalize : SearchWordsとExcludeWordsを合わせる
// SearchWordsは小文字にする
// ExcludeWordsは小文字にした上で
// ソートして、頭に-をつける
func Normalize(se, ex []string) string {
	// Sort
	sort.Slice(ex, func(i, j int) bool { return ex[i] < ex[j] })
	// Add prefix "-"
	strs := append(se, func() (d []string) {
		for _, ex := range ex {
			d = append(d, "-"+ex)
		}
		return
	}()...)
	return strings.Join(strs, " ")
}

// ResolveLocateCmd は使用するlocateコマンドのパスを解決する。
// 優先順位:
//  1. 引数 explicit (フラグ or 環境変数 LOCATE_CMD で明示指定)
//  2. PATH上の gocate
//  3. PATH上の locate
func ResolveLocateCmd(explicit string) (string, error) {
	if explicit != "" {
		if _, err := exec.LookPath(explicit); err != nil {
			return "", fmt.Errorf("LOCATE_CMD %q not found: %w", explicit, err)
		}
		return explicit, nil
	}
	for _, candidate := range []string{"gocate", "locate"} {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no locate command found (tried: gocate, locate)")
}
