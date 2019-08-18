package locater

import "testing"

func TestRuneIndex(t *testing.T) {
	s := "あtTの5\\UW"
	actual := RuneIndex(s, '\\')
	expect := 5
	if expect != actual {
		t.Fatalf("got: %v want: %v", actual, expect)
	}
}

func TestToLowerExcept(t *testing.T) {
	s := "あtTの5\\UW"
	actual := ToLowerExcept(s, '\\')
	expect := "あttの5\\Uw"
	if expect != actual {
		t.Fatalf("got: %v want: %v", actual, expect)
	}
}

func TestToLowerExceptFirst(t *testing.T) {
	s := "SあtT5\\Uほw\\dHo\\T"
	actual := ToLowerExceptFirst(s)
	expect := "Sあtt5\\uほw\\dho\\t"
	if expect != actual {
		t.Fatalf("got: %v want: %v", actual, expect)
	}
}

func TestToLowerExceptAll(t *testing.T) {
	s := "\\SあtT5\\Uほw\\dHo\\T"
	actual := ToLowerExceptAll(s, '\\')
	expect := "\\Sあtt5\\Uほw\\dho\\T"
	if expect != actual {
		t.Fatalf("got: %v want: %v", actual, expect)
	}
}

func TestQueryParserError_Test(t *testing.T) {
	receiveValue := "Dropbox -012 34 Program PYTHON -Go -Joker -99 as\\Wgo -ab\\D\\d"
	se := []string{"dropbox", "34", "program", "python", "as\\Wgo"} // バックスラッシュ後はlowerしない
	ex := []string{"012", "go", "joker", "99", "ab\\D\\d"}          // QueryParserはsortしない
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

	receiveValue = "a b c"
	if _, _, err := QueryParser(receiveValue); err == nil {
		t.Errorf("This test must fail")
	}

	receiveValue = "dropbox Program VIM -Vim"
	if _, _, err := QueryParser(receiveValue); err == nil {
		t.Errorf("This test must fail")
	}
}
