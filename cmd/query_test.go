package locater

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

// func TestQuery_LimitParser(t *testing.T) {
// 	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
// 	req, _ := http.NewRequest("GET", "/json?limit=25", nil)
// 	ginContext.Request = req
//
// 	actual := ginContext.GetInt("limit")
// 	expect := 24
// 	if expect != actual {
// 		t.Fatalf("got: %v want: %v", actual, expect)
// 	}
// }

func TestQuery_Parser2_Test(t *testing.T) {
	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	req, _ := http.NewRequest("GET", "/json?q=search+go&logging=false&limit=25", nil)
	ginContext.Request = req

	q := Query{}
	err := ginContext.ShouldBind(&q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	actual := q
	expected := Query{
		Q:       "search go",
		Logging: false,
		Limit:   25,
	}
	if actual.Q != expected.Q {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
	if actual.Logging != expected.Logging {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
	if actual.Limit != expected.Limit {
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
