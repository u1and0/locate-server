package locater

import (
	"testing"
)

func TestLocater_Normalize(t *testing.T) {
	// QueryParserによってSearchWordsとExcludeWordsは小文字に正規化されている
	l := Locater{
		SearchWords:  []string{"dropbox", "program", "34"},        // Should be lower
		ExcludeWords: []string{"543", "python", "12", "go", "漢字"}, // Should be sort & lower
	}

	actual := l.Normalize()
	expected := "dropbox program 34 -12 -543 -go -python -漢字"

	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
