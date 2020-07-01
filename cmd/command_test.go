package locater

import (
	"strings"
	"testing"
)

func TestLocateStats(t *testing.T) {
	// os.Setenv("LOCATE_PATH", "../test/mlocatetest.db:../test/mlocatetest1.db")
	b, _ := LocateStats("../test/mlocatetest.db:../test/mlocatetest1.db")
	actual := strings.Fields(string(b))
	expected := strings.Fields(`
	データベース ../test/mlocatetest.db:
		67 辞書
		1,155 ファイル
		ファイル名に 43,839 バイト
		データベースに保存するのに 23,198 バイト使いました
	データベース ../test/mlocatetest1.db:
		5,581 辞書
		73,710 ファイル
		ファイル名に 3,294,594 バイト
		データベースに保存するのに 1,600,378 バイト使いました
	`)
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("got: %v want: %v", actual[i], expected[i])
		}
	}
}

func TestLocateStatsSum(t *testing.T) {
	// os.Setenv("LOCATE_PATH", "../test/mlocatetest.db:../test/mlocatetest1.db")
	b, _ := LocateStats("../test/mlocatetest.db:../test/mlocatetest1.db")
	actual, _ := LocateStatsSum(b)
	expected := uint64(74865)
	if actual != expected {
		t.Fatalf("got: %d want: %d\n$ locate -S\n%v", actual, expected, string(b))
	}
}

func Test_Ambiguous(t *testing.T) {
	actual := []uint64{1000000000, 100000000, 1999999, 2345678, 30001, 4021, 56}
	expected := []string{"10億", "1億", "1百万", "2百万", "3万", "4千", "56"}
	for i, a := range actual {
		ag := Ambiguous(a)
		if ag != expected[i] {
			t.Fatalf("got: %s want: %s", ag, expected[i])
		}
	}
}

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
	// Single process test
	l := Locater{
		Process:      1,
		SearchWords:  []string{"the", "path", "for", "search"},
		ExcludeWords: []string{"exclude", "paths"},
		Dbpath:       "/var/lib/mlocate/mlocatetest.db",
	}
	actual := l.CmdGen()
	expected := [][]string{
		[]string{"locate", "--ignore-case", "--quiet",
			"--regex", "the.*path.*for.*search",
			"--database", "/var/lib/mlocate/mlocatetest.db"},
		[]string{"grep", "-ivE", "exclude"},
		[]string{"grep", "-ivE", "paths"},
	}
	t.Logf("ex: %v, ac: %v", expected, actual) // Print command
	for i, e1 := range expected {
		for j, e2 := range e1 {
			if actual[i][j] != e2 {
				t.Fatalf("got: %v want: %v\ncommand: %s", actual[i][j], e2, actual)
			}
		}
	}

	// Multi process test
	l = Locater{
		Process:      0,
		SearchWords:  []string{"the", "path", "for", "search"},
		ExcludeWords: []string{"exclude", "paths"},
		Dbpath:       "../test/mlocatetest.db:../test/mlocatetest1.db",
	}
	actual = l.CmdGen()
	expected = [][]string{
		[]string{
			"locate", "--ignore-case", "--quiet",
			"--regex", "the.*path.*for.*search",
			"--database", "/var/lib/mlocate/mlocatetest.db",
		},
		[]string{"grep", "-ivE", "exclude"},
		[]string{"grep", "-ivE", "paths"},
		}
}
