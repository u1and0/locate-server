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

func TestQuery_Parser_Test(t *testing.T) {
	var (
		ginContext, _ = gin.CreateTestContext(httptest.NewRecorder())
		err           error
		query         Query
	)

	/* Succsess test */
	// Typical URL request
	req, _ := http.NewRequest("GET", "/json?q=search+Go&logging=false&limit=25", nil)
	ginContext.Request = req
	actual := query.New()
	if err = ginContext.ShouldBind(&actual); err != nil {
		t.Fatalf("error: %#v", err)
	}
	expected := Query{
		Q:       "search Go",
		Logging: false,
		Limit:   25,
	}
	if actual.Q != expected.Q {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}
	if actual.Logging != expected.Logging {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}
	if actual.Limit != expected.Limit {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}

	// Large Case
	req, _ = http.NewRequest("GET", "/json?q=&logging=TRUE", nil)
	ginContext.Request = req
	actual = query.New()
	if err = ginContext.ShouldBind(&actual); err != nil {
		t.Fatalf("error: %#v", err)
	}
	expected = Query{
		Q:       "",
		Logging: true,
		Limit:   0,
	}
	if actual.Q != expected.Q {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}
	if actual.Logging != expected.Logging {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}
	if actual.Limit != expected.Limit {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}

	// Ommit query
	// Default value is q="", logging=false, limit=0
	// q=="" => Should Error
	// logging==false => Should be true
	// limit==0 => Should be -1
	req, _ = http.NewRequest("GET", "/json?q=", nil)
	ginContext.Request = req
	actual = query.New()
	if err = ginContext.ShouldBind(&actual); err != nil {
		t.Fatalf("error: %#v", err)
	}
	expected = Query{
		Q:       "",
		Logging: true,
		Limit:   -1,
	}
	if actual.Q != expected.Q {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}
	if actual.Logging != expected.Logging {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}
	if actual.Limit != expected.Limit {
		t.Fatalf("got: %#v want: %#v", actual, expected)
	}

	/* Error test */
	// Int out of range
	req, _ = http.NewRequest("GET", "/json?q=search+Go&limit=0.1", nil)
	ginContext.Request = req
	actual = query.New()
	if err = ginContext.ShouldBind(&actual); err == nil {
		t.Errorf(
			"This test must fail by %s",
			`error: &strconv.NumError{Func:"ParseInt", Num:"0.1", Err:(*errors.errorString)`,
		)
	}

	// Invalid boolian
	req, _ = http.NewRequest("GET", "/json?q=search+Go&logging=hoge", nil)
	ginContext.Request = req
	actual = query.New()
	if err = ginContext.ShouldBind(&actual); err == nil {
		t.Errorf(
			"This test must fail by %s",
			`error: &strconv.NumError{Func:"ParseBool", Num:"hoge", Err:(*errors.errorString)`,
		)
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
