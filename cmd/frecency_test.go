package locater

import (
	"strings"
	"testing"
	"time"
)

func Test_LogWord(t *testing.T) {
	expected := History{
		"load bash": []time.Time{
			time.Date(2020, 6, 28, 20, 59, 13, 0, time.UTC),
		},
		"yay package": []time.Time{
			time.Date(2020, 6, 28, 20, 59, 33, 0, time.UTC),
		},
		"etc pacman new": []time.Time{
			time.Date(2020, 6, 28, 21, 35, 15, 0, time.UTC),
			time.Date(2020, 6, 28, 21, 41, 25, 0, time.UTC),
			time.Date(2020, 6, 28, 21, 42, 36, 0, time.UTC),
			time.Date(2020, 6, 28, 22, 28, 16, 0, time.UTC),
		},
		"usr pac": []time.Time{
			time.Date(2020, 9, 27, 7, 39, 46, 0, time.UTC),
			time.Date(2020, 9, 27, 7, 46, 54, 0, time.UTC),
			time.Date(2020, 9, 27, 7, 47, 05, 0, time.UTC),
		},
	}
	actual, err := LogWord("../test/locate.log")
	if err != nil {
		t.Errorf("error: %v", err)
	}
	for k, v := range actual {
		// length of map value test
		if len(expected[k]) != len(v) {
			t.Fatalf("got: %v want: %v", actual, expected)
		}
		// time format test
		ex := expected[k][0]
		if ex != v[0] {
			t.Fatalf("got: %v want: %v", v[0], ex)
		}
	}
}

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

func Test_ExtractKeyword(t *testing.T) {
	expected := "usr pac"
	actual := strings.TrimSpace(ExtractKeyword(`
[32m[NOTICE] ▶ 2020-09-27 07:46:54.418 main.go:263     2666files 144.478msec PUSH result to cache [ usr pac                                            ] [0m
`))
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

func Test_Datalist(t *testing.T) {
	f := FrecencyList{Frecency{"hoge", 1}, Frecency{"foo", 0}}
	actual := f.Datalist()
	expected := `<option value="hoge"></option><option value="foo"></option>`
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
