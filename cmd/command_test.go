package locater

import (
	"os"
	"strings"
	"testing"
)

func TestLocateStats(t *testing.T) {
	dbpath := "../test/etc.db:../test/root.db:../test/usr.db"
	b, _ := LocateStats(dbpath)
	actual := strings.Fields(string(b))
	expected := strings.Fields(`
		データベース ../test/etc.db:
		67 辞書
		1,155 ファイル
		ファイル名に 43,839 バイト
		データベースに保存するのに 23,198 バイト使いました
		データベース ../test/root.db:
		316 辞書
		1,772 ファイル
		ファイル名に 140,093 バイト
		データベースに保存するのに 111,148 バイト使いました
		データベース ../test/usr.db:
		5,581 辞書
		73,710 ファイル
		ファイル名に 3,294,594 バイト
		データベースに保存するのに 1,600,378 バイト使いました
	`)
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("got: %v want: %v\n%s",
				actual[i], expected[i], "このテストは/var/lib/mlocate.dbがあると失敗する")
		}
	}
}

func TestLocateStatsSum(t *testing.T) {
	if _, err := os.Stat("/var/lib/mlocate/mlocate.db"); err == nil {
		t.Fatal("このテストは/var/lib/mlocate/mlocate.dbがあると失敗する")
	}
	dbpath := "../test/etc.db:../test/root.db:../test/usr.db"
	b, err := LocateStats(dbpath)
	if err != nil {
		t.Fatalf("LocateStats error occur %s", err)
	}
	actual, err := LocateStatsSum(b)
	if err != nil {
		t.Fatalf("LocateStatsSum error occur %s", err)
	}
	expected := uint64(76637)
	if actual != expected {
		t.Fatalf("got: %v want: %v\n$ locate -S\n%v\n",
			actual, expected, string(b))
	}
}

func Test_Ambiguous(t *testing.T) {
	actual := []uint64{1000000000, 100000000, 1999999, 2345678, 30001, 4021, 56}
	expected := []string{"10億", "1億", "199万", "234万", "3万", "4千", "56"}
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
		Dbpath:       "../test/mlocatetest.db:../test/mlocatetest1.db",
	}
	actual := l.CmdGen()
	expected := [][]string{
		[]string{"locate", "--ignore-case", "--quiet",
			"--regex", "the.*path.*for.*search",
			"--database", "../test/mlocatetest.db:../test/mlocatetest1.db"},
		[]string{"grep", "-ivE", "exclude"},
		[]string{"grep", "-ivE", "paths"},
	}
	t.Logf("expected command: %v, actual command: %v", expected, actual) // Print command
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
		[]string{"echo", "../test/mlocatetest.db:../test/mlocatetest1.db"},
		[]string{"tr", ":", "'\\n'"},
		[]string{"xargs", "-P", "0", "-I@",
			"locate", "--ignore-case", "--quiet",
			"--regex", "the.*path.*for.*search",
			"--database", "@"},
		[]string{"grep", "-ivE", "exclude"},
		[]string{"grep", "-ivE", "paths"},
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
