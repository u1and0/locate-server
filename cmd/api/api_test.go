package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHistoryQueryParser(t *testing.T) {
	var ginContext, _ = gin.CreateTestContext(httptest.NewRecorder())
	req, _ := http.NewRequest("GET", "/history?gt=10&lt=100", nil)
	ginContext.Request = req

	actual, err := HistoryQueryParser(ginContext, "gt")
	if err != nil {
		t.Fatalf("%v, %v", err, ginContext.Request)
	}
	expected := 10
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}

	actual, err = HistoryQueryParser(ginContext, "lt")
	if err != nil {
		t.Fatalf("%v, %v", err, ginContext.Request)
	}
	expected = 100
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}
}
