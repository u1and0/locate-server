package locater

import (
	"reflect"
	"testing"
)

func TestLocater_CmdGen(t *testing.T) {
	l := Locater{
		SearchWords:  []string{"the", "path", "for", "search"},
		ExcludeWords: []string{"exclude", "paths"},
		Args:         Args{Dbpath: "../test"},
	}

	actual := l.CmdGen()

	expected := [][]string{
		{
			"gocate",
			"--database",
			"../test",
			"--",
			"--ignore-case",
			"--existing",
			"--regex",
			"the.*path.*for.*search",
		},
		{"grep", "-ivE", "exclude"},
		{"grep", "-ivE", "paths"},
	}

	// Compare slices slice-by-slice
	for i, expCmd := range expected {
		if !reflect.DeepEqual(actual[i], expCmd) {
			t.Fatalf("cmd %d: got %v, want %v", i, actual[i], expCmd)
		}
	}
}
