package locater

import (
	"testing"
)

func TestLocateStats(t *testing.T) {
	actual, _ := LocateStats("../test")
	var expected int64
	expected = 111155
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

func Test_Ambiguous(t *testing.T) {
	actual := []int64{1_100_000_000, 205_000_000, 3_999_999, 434_567, 50_001, 6_021, 783, 86}
	expected := []string{"1,000,000,000+", "205,000,000+", "3,000,000+", "434,000+", "50,000+", "6,000+", "783", "86"}
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
