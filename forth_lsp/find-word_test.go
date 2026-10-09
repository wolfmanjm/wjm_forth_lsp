package forth_lsp

import "testing"

func TestWord(t *testing.T) {
	str := "This is a line of text"

	got := WordAtPoint(str, 0)
    want := "This"
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 0 %s; want %s", got, want)
    }

	got = WordAtPoint(str, 4)
    want = ""
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 4 '%s'; want '%s'", got, want)
    }

	got = WordAtPoint(str, 8)
    want = "a"
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 8 '%s'; want '%s'", got, want)
    }

	got = WordAtPoint(str, 13)
    want = "line"
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 13 '%s'; want '%s'", got, want)
    }

	got = WordAtPoint(str, 18)
    want = "text"
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 18 '%s'; want '%s'", got, want)
    }

	got = WordAtPoint(str, 21)
    want = "text"
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 18 '%s'; want '%s'", got, want)
    }

	got = WordAtPoint(str, 22)
    want = ""
    if got != want {
        // t.Errorf logs the error and allows the test suite to keep running
        t.Errorf("at point 22 '%s'; want '%s'", got, want)
    }
}

func TestTableDriven(t *testing.T) {
	str := "This is a line of (blah) text"
    // Define the test cases
    tests := []struct {
        name string
        pos int
        want string
    }{
        {"pos 0", 0, "This"},
        {"pos 6", 6, "is"},
        {"pos 7", 7, ""},
        {"pos 8", 8, "a"},
        {"pos 20", 20, "(blah)"},
        {"pos 21", 28, "text"},
        {"pos 22", 29, ""},
    }

    // Iterate and run each case as a subtest
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := WordAtPoint(str, tt.pos)
            if got != tt.want {
                t.Errorf("WordAtPoint(%s, %d) = %s; want %s", str, tt.pos, got, tt.want)
            }
        })
    }
}
