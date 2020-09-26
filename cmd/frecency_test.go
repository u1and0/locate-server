package locater

import (
	"testing"
	"time"
)

func Test_ExtractDatetime(t *testing.T) {
	s := `
[32m[NOTICE] ▶ 2020-07-07 06:57:27.667 main.go:233        0files 4.607msec PUSH result to cache [ usr                                                ] [0m
	`
	layout := "2006-01-02 15:04:05"
	expected, err := time.Parse(layout, "2020-07-07 06:57:27")
	actual, err := ExtractDatetime(s)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

/*
func Test_LogWord(t *testing.T){
	actual := LogWord("test/locate.log")
	expected := map[]
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
*/

// func Test_Scoring(t *testing.T) {
// 	now := time.Now()
// 	tt := []time.Time{
// 		now.Sub(5*time.Hour),
// 	}
// 	for _,t := range tt{
// 		actual := Scoring(t)
// 		expected := 16
// 		if actual != expected{
// 			t.Fatalf("got: %v want: %v", actual, expected)
// 		}
// 	}
