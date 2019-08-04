package locater

import (
	"testing"
)

func TestLocater_CmdGen(t *testing.T) {
	l := Locater{
		SearchWords:  []string{"the", "path", "for", "search"},
		ExcludeWords: []string{"exclude", "paths"},
	}
	actual := l.CmdGen()
	expected := [][]string{
		[]string{"locate", "-i", "--regex", "the.*path.*for.*search"},
		[]string{"grep", "-ivE", "exclude"},
		[]string{"grep", "-ivE", "paths"},
	}

	for i, e1 := range expected {
		for j, e2 := range e1 {
			if actual[i][j] != e2 {
				t.Fatalf("got: %v want: %v", actual[i][j], e2)
			}

		}
	}
}
