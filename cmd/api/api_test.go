package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHistoryQueryParser(t *testing.T) {
	var ginContext, _ = gin.CreateTestContext(httptest.NewRecorder())

	req, _ := http.NewRequest("GET", "/history?lt=10", nil)
	ginContext.Request = req
	actual, err := HistoryQueryParser(ginContext, "lt")
	if err != nil {
		t.Fatalf("%v, %v", err, ginContext.Request)
	}
	expected := 10
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}

	var gginContext, _ = gin.CreateTestContext(httptest.NewRecorder())
	req, _ = http.NewRequest("GET", "/history?gt=10", nil)
	gginContext.Request = req
	actual, err = HistoryQueryParser(gginContext, "gt")
	if err != nil {
		t.Fatalf("%v, %v", err, gginContext.Request)
	}
	expected = 10
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, gginContext.Request)
	}

}
