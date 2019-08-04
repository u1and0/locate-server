package locater

import (
	"testing"
)

func TestQueryParserError_Test(t *testing.T) {
	receiveValue := "Dropbox -012 34 Program PYTHON -Go -Joker -99"
	se := []string{"dropbox", "34", "program", "python"}
	ex := []string{"012", "go", "joker", "99"} // QueryParserはsortしない
	ase, aex, _ := QueryParser(receiveValue)

	for i, s := range se {
		if ase[i] != s {
			t.Fatalf("got: %v want: %v", ase, se)
		}
	}

	for i, e := range ex {
		if aex[i] != e {
			t.Fatalf("got: %v want: %v", aex, ex)
		}
	}

	receiveValue = "a"
	if _, _, err := QueryParser(receiveValue); err == nil {
		t.Errorf("This test must fail")
	}

	receiveValue = "dropbox Program VIM -Vim"
	if _, _, err := QueryParser(receiveValue); err == nil {
		t.Errorf("This test must fail")
	}
}
