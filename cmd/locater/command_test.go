package locater

import (
	"testing"
)

func TestDBSize(t *testing.T) {
	// 未実装
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
