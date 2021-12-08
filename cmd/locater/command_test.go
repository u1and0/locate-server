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
	b, err := LocateStats("../../test")
	if err != nil {
		t.Fatalf("LocateStats error occur %s", err)
	}
	actual, err := LocateStatsSum(b)
	if err != nil {
		t.Fatalf("LocateStatsSum error occur %s", err)
	}
	expected := 76637
	if actual != expected {
		t.Fatalf("got: %v want: %v\n$ locate -S\n%v\n",
			actual, expected, string(b))
	}
}

func Test_Ambiguous(t *testing.T) {
	actual := []int{1_100_000_000, 200_000_000, 3_999_999, 434_567, 50_001, 6_021, 783, 86}
	expected := []string{"1000000000", "200000000", "3000000", "400000", "50000", "6000", "783", "86"}
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
