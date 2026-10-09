package forth_lsp

import (
	"context"
	"log"
	"strings"

	"github.com/owenrumney/go-lsp/document"
	"github.com/owenrumney/go-lsp/lsp"
	"github.com/owenrumney/go-lsp/server"
)

// Handler implements the LSP handler interfaces for mylang.
type Handler struct {
	documents *document.Store
	client *server.Client
}

// New creates a new Handler.
func New() *Handler {
	return &Handler{
		documents: document.NewStore(),
	}
}

func (h *Handler) SetClient(client *server.Client) {
	h.client = client
}

func (h *Handler) Initialize(_ context.Context, _ *lsp.InitializeParams) (*lsp.InitializeResult, error) {
	err := OpenDatabase()
	if err != nil {
		log.Fatalf("Database failed: %v", err)
	}
	return &lsp.InitializeResult{
		ServerInfo: &lsp.ServerInfo{Name: "mylang", Version: "0.1.0"},
	}, nil
}

func (h *Handler) Shutdown(_ context.Context) error {
	CloseDatabase()
	return nil
}

func (h *Handler) DidOpen(_ context.Context, params *lsp.DidOpenTextDocumentParams) error {
	_, err := h.documents.Open(params)
	return err
}

func (h *Handler) DidChange(_ context.Context, params *lsp.DidChangeTextDocumentParams) error {
	_, err := h.documents.Change(params)
	return err
}

func (h *Handler) DidClose(_ context.Context, params *lsp.DidCloseTextDocumentParams) error {
	h.documents.Close(params)
	return nil
}

func (h *Handler) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) error {
	var diags []lsp.Diagnostic

	text, ok := h.documents.Text(params.TextDocument.URI)
	if ok {
		for i, line := range strings.Split(text, "\n") {
			idx := strings.Index(line, "\\ TODO")
			if idx < 0 {
				continue
			}
			sev := lsp.SeverityWarning
			diags = append(diags, lsp.Diagnostic{
				Range: lsp.Range{
					Start: lsp.Position{Line: i, Character: idx},
					End:   lsp.Position{Line: i, Character: idx + 4},
				},
				Severity: &sev,
				Source:   "forth",
				Message:  "TODO found",
			})
		}
	}

	return h.client.PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: diags,
	})
}
