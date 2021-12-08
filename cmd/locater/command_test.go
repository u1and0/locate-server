package locater

import (
	"strings"
	"testing"
)

func TestLocateStats(t *testing.T) {
	b, _ := LocateStats("../test")
	actual := strings.Fields(string(b))
	expected := strings.Fields(`
		データベース ../test/etc.db:
		67 辞書
		1,155 ファイル
		ファイル名に 43,839 バイト
		データベースに保存するのに 23,198 バイト使いました
		データベース ../test/root.db:
		316 辞書
		1,772 ファイル
		ファイル名に 140,093 バイト
		データベースに保存するのに 111,148 バイト使いました
		データベース ../test/usr.db:
		5,581 辞書
		73,710 ファイル
		ファイル名に 3,294,594 バイト
		データベースに保存するのに 1,600,378 バイト使いました
	`)
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("got: %v want: %v\n%s",
				actual[i], expected[i], "このテストは/var/lib/mlocate.dbがあると失敗する")
		}
	}
}

func TestLocateStatsSum(t *testing.T) {
	b, err := LocateStats("../test")
	if err != nil {
		t.Fatalf("LocateStats error occur %s", err)
	}
	actual, err := LocateStatsSum(b)
	if err != nil {
		t.Fatalf("LocateStatsSum error occur %s", err)
	}
	expected := uint64(76637)
	if actual != expected {
		t.Fatalf("got: %v want: %v\n$ locate -S\n%v\n",
			actual, expected, string(b))
	}
}

func Test_Ambiguous(t *testing.T) {
	actual := []uint64{1000000000, 100000000, 1999999, 2345678, 30001, 4021, 56}
	expected := []string{"10億", "1億", "199万", "234万", "3万", "4千", "56"}
	for i, a := range actual {
		ag := Ambiguous(a)
		if ag != expected[i] {
			t.Fatalf("got: %s want: %s", ag, expected[i])
		}
	}
}

func TestNormalize(t *testing.T) {
	// QueryParserによってSearchWordsとExcludeWordsは小文字に正規化されている
	sw := []string{"dropbox", "program", "34"}        // Should be lower
	ew := []string{"543", "python", "12", "go", "漢字"} // Should be sort & lower

	actual := Normalize(sw, ew)
	expected := "dropbox program 34 -12 -543 -go -python -漢字"

	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
