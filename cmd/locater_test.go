package locater

import (
	"testing"
)

func TestLocater_Normalize(t *testing.T) {
	// QueryParserによってSearchWordsとExcludeWordsは小文字に正規化されている
	l := Locater{
		Query: Query{
			SearchWords:  []string{"dropbox", "program", "34"},        // Should be lower
			ExcludeWords: []string{"543", "python", "12", "go", "漢字"}, // Should be sort & lower
		},
	}

	actual := l.Normalize()
	expected := "dropbox program 34 -12 -543 -go -python -漢字"

	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

func TestLocater_CmdGen(t *testing.T) {
	l := Locater{
		Query: Query{
			SearchWords:  []string{"the", "path", "for", "search"},
			ExcludeWords: []string{"exclude", "paths"},
		},
		Args: Args{Dbpath: "../test"},
	}
	actual := l.CmdGen()
	expected := [][]string{
		{
			"gocate",
			"--database",
			"../test",
			"--",
			"--ignore-case",
			"--quiet",
			"--existing",
			"--nofollow",
			"--regex",
			"the.*path.*for.*search",
		},
		{"grep", "-ivE", "exclude"},
		{"grep", "-ivE", "paths"},
	}
	t.Logf("expected command: %v, actual command: %v", expected, actual) // Print command
	for i, e1 := range expected {
		for j, e2 := range e1 {
			if actual[i][j] != e2 {
				t.Fatalf("got: %v want: %v\ncommand: %s", actual[i][j], e2, actual)
			}
		}
	}
}
