package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntQuery(t *testing.T) {
	var ginContext, _ = gin.CreateTestContext(httptest.NewRecorder())
	req, _ := http.NewRequest("GET", "/history?gt=10&lt=100", nil)
	ginContext.Request = req

	// request gt,lt
	actual := IntQuery(ginContext, "gt")
	expected := 10
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}

	actual = IntQuery(ginContext, "lt")
	expected = 100
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}

	// no request gt,lt
	ginContext, _ = gin.CreateTestContext(httptest.NewRecorder())
	req, _ = http.NewRequest("GET", "/history", nil)
	ginContext.Request = req

	actual = IntQuery(ginContext, "gt")
	expected = 0
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}
	actual = IntQuery(ginContext, "lt")
	expected = 0
	if actual != expected {
		t.Fatalf("got: %v want: %v, %v", actual, expected, ginContext.Request)
	}

}
