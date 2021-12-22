package api

import (
	"testing"
)

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

func Test_QueryParser(t *testing.T) {
	testWord := `hoGe my -Your name\D -HeY`
	sn, en, _ := QueryParser(testWord)
	expectedSN := []string{"hoge", "my", `name\D`}
	expectedEN := []string{"your", "hey"}
	for i, actual := range sn {
		if expectedSN[i] != actual {
			t.Fatalf("got: %v want: %v", sn, expectedSN)
		}
	}
	for i, actual := range en {
		if expectedEN[i] != actual {
			t.Fatalf("got: %v want: %v", en, expectedEN)
		}
	}
}
func Test_lessInput(t *testing.T) {
	// lessInput case
	expected := true

	testWord := []string{"a", "b"}
	actual := lessInput(testWord)
	if actual != expected {
		t.Errorf("testWord: %v got: %v want: %v", testWord, actual, expected)
	}
	testWord = []string{"\a", "b"}
	actual = lessInput(testWord)
	if actual != expected {
		t.Errorf("testWord: %v got: %v want: %v", testWord, actual, expected)
	}

	// NOT lessInput case
	expected = false

	testWord = []string{"a", "bb"}
	actual = lessInput(testWord)
	if actual != expected {
		t.Errorf("testWord: %v got: %v want: %v", testWord, actual, expected)
	}
	testWord = []string{"\a", "\bc"}
	actual = lessInput(testWord)
	if actual != expected {
		t.Errorf("testWord: %v got: %v want: %v", testWord, actual, expected)
	}
}

func Test_duplicateWord(t *testing.T) {
	sn := []string{"ahoy", "book"}
	en := []string{"ahoy"}
	actual := duplicateWord(sn, en)
	expected := "ahoy"
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

// func TestQuery_ParserError_Test(t *testing.T) {
// 	receiveValue := "Dropbox -012 34 Program PYTHON -Go -Joker -99 \\Tas\\Wgo -ab\\D\\d"
// 	se := []string{"dropbox", "34", "program", "python", "\\Tas\\Wgo"} // バックスラッシュ後はlowerしない
// 	ex := []string{"012", "go", "joker", "99", "ab\\D\\d"}             // QueryParserはsortしない
// 	query := Query{se, ex, false, 0}
// 	ase, aex, _ := query.Parser(receiveValue)
//
// 	for i, s := range se {
// 		if ase[i] != s {
// 			t.Fatalf("got: %v want: %v", ase, se)
// 		}
// 	}
//
// 	for i, e := range ex {
// 		if aex[i] != e {
// 			t.Fatalf("got: %v want: %v", aex, ex)
// 		}
// 	}
//
// 	receiveValue = "a"
// 	if _, _, err := query.Parser(receiveValue); err == nil {
// 		t.Errorf("This test must fail")
// 	}
//
// 	receiveValue = "a b c"
// 	if _, _, err := query.Parser(receiveValue); err == nil {
// 		t.Errorf("This test must fail")
// 	}
//
// 	receiveValue = "dropbox Program VIM -Vim"
// 	if _, _, err := query.Parser(receiveValue); err == nil {
// 		t.Errorf("This test must fail")
// 	}
// }
