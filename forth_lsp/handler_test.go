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
	fmt.Fprintln(os.Stderr, hover.Contents.Markup.Value)
	// if hover.Contents.Markup.Value != "**mylang** hover: there" {
	// 	t.Error("expected hover not found")
	// }

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

	h.DidOpen("file:///test.txt", "mylang", "hello")

	result, err := h.Completion("file:///test.txt", 0, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) == 0 {
		t.Error("expected at least one completion item")
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
