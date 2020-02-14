package locater

import "testing"

func TestLocater_highlightString(t *testing.T) {
	s := "/home/vagrant/Program/hoge3/program_boot.pdf"
	words := []string{"program", "pdf"}
	actual := highlightString(s, words)
	p := "<span style=\"background-color:#FFCC00;\">"
	q := "</span>"
	expected := "/home/vagrant/" +
		p + "Program" + q +
		"/hoge3/program_boot." +
		p + "pdf" + q
	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}

func TestLocater_CmdGen(t *testing.T) {
	l := Locater{
		SearchWords:  []string{"the", "path", "for", "search"},
		ExcludeWords: []string{"exclude", "paths"},
		Dbpath:       "/var/lib/mlocate/mlocatetest.db",
	}
	actual := l.CmdGen()
	expected := [][]string{
		[]string{"locate", "-i",
			"-d", "/var/lib/mlocate/mlocatetest.db",
			"--regex", "the.*path.*for.*search"},
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
