package forth_lsp_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/owenrumney/go-lsp/servertest"
	"github.com/wolfmanjm/wjm_forth_lsp/forth_lsp"
)

func TestInitialize(t *testing.T) {
	h := servertest.New(t, forth_lsp.New())

	if h.InitResult.ServerInfo == nil || h.InitResult.ServerInfo.Name != "mylang" {
		t.Error("expected server info to be set")
	}
}

func TestHover(t *testing.T) {
	h := servertest.New(t, forth_lsp.New())

	h.DidOpen("file:///test.txt", "mylang", "fred\nud/mod\n")

	hover, err := h.Hover("file:///test.txt", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if hover == nil {
		t.Fatal("expected hover result")
	}

	if hover.Contents.Markup.Value != "*ud/mod*  `( ud1 ud2 -- ud3 ud4 )`  \n64/64 = 64 rem 64 Division  \n"  {
		fmt.Fprintln(os.Stderr, hover.Contents.Markup.Value)
		t.Error("expected hover not found")
	}

	hover, err = h.Hover("file:///test.txt", 0, 3)
	if err != nil {
		t.Fatal(err)
	}
	if hover != nil {
		fmt.Fprintln(os.Stderr, hover.Contents.Markup.Value)
		t.Fatal("did not expect hover result")
	}
}

func TestCompletion(t *testing.T) {
	h := servertest.New(t, forth_lsp.New())

	h.DidOpen("file:///test.txt", "mylang", "du")

	result, err := h.Completion("file:///test.txt", 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 4 {
		t.Error("expected 4 completion items, got: ", len(result.Items))
	}

	expected := []string {
		"du<",
		"du>",
		"dump",
		"dup",
	}
	for i, v := range result.Items {
		if v.Label != expected[i] {
			t.Error("expected: ", expected[i], " got: ", v.Label)
		}
	}
}

func TestDiagnostics(t *testing.T) {
	h := servertest.New(t, forth_lsp.New())

	h.DidOpen("file:///test.txt", "mylang", "\\ TODO fix this")
	h.DidSave("file:///test.txt")

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	diags, err := h.WaitForDiagnostics(ctx, "file:///test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(diags) == 0 {
		t.Error("expected diagnostics for TODO")
	}
}
