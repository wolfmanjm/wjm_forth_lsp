package main

import (
	_ "fmt"
	"testing"
)

func TestTableDriven(t *testing.T) {
    // Define the test cases
    tests := []struct {
        name string
        src string
        want Word
    }{
        {
        	"defn",
        	"\\ this is doc\n: test ( a -- b ) blah ;",
        	Word{"test", ":", 2,  "a -- b", "this is doc", false, false},
        },
        {
        	"variable",
        	"\\ this is doc\n23 variable testvar  \\ more doc",
        	Word{"testvar", "VARIABLE", 2,  "", "this is doc\nmore doc", false, false},
        },
        {
        	"buffer",
        	"\\ this is doc\n23 buffer: buf1",
        	Word{"buf1", "BUFFER:", 2,  "", "this is doc", false, false},
        },

    }

    // Iterate and run each case as a subtest
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            words := parse(tt.src, false, nil)
            if len(words) != 1 {
            	 t.Errorf("Expected 1 word, got %v", len(words))
            }
            // fmt.Println(len(words), words[0])
            if *words[0] != tt.want {
                t.Errorf("Expected %v, got %v", tt.want, *words[0])
            }
        })
    }
}
