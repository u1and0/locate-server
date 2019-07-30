package locater

import (
	"testing"
)

func TestLocater_Normalize(t *testing.T) {
	l := Locater{
		SearchWords:  []string{"DropBox", "Program", "34"},        // Should be lower
		ExcludeWords: []string{"543", "PYTHON", "12", "go", "漢字"}, // Should be sort & lower
	}

	actual := l.Normalize()
	expected := "dropbox program 34 -12 -543 -go -python -漢字"

	if actual != expected {
		t.Fatalf("got: %v want: %v", actual, expected)
	}
}
